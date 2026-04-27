import { Request, Response, NextFunction } from 'express';
import { ZodError, ZodSchema } from 'zod';

export const validateRequest = (schema: ZodSchema<any>) => {
    return async (req: Request, res: Response, next: NextFunction) => {
        try {
            const validatedData = await schema.parseAsync({
                body: req.body,
                query: req.query,
                params: req.params,
            });

            // ক্লিন করা ডেটা আবার রিকোয়েস্টে সেট করা
            req.body = validatedData.body;
            req.query = validatedData.query;
            req.params = validatedData.params;

            next();
        } catch (error: any) {
            if (error instanceof ZodError) {
                // ZodError থেকে সরাসরি ইস্যুগুলো ম্যাপ করা
                const formattedErrors = error.issues.map((issue) => {
                    return {
                        // পাথ যদি ['body', 'email'] হয়, তবে আমরা 'email' নেব
                        path: issue.path.length > 0 ? issue.path[issue.path.length - 1] : 'request',
                        message: issue.message,
                    };
                });

                return res.status(400).json({
                    success: false,
                    message: 'Validation Error',
                    errors: formattedErrors, // এখন এটি আর খালি আসবে না
                });
            }
            next(error);
        }
    };
};