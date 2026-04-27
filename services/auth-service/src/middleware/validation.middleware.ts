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

            req.body = validatedData.body;
            req.query = validatedData.query;
            req.params = validatedData.params;

            next();
        } catch (error: any) {
            if (error instanceof ZodError) {
                const formattedErrors = error.issues.map((issue) => {
                    return {
                        path: issue.path.length > 0 ? issue.path[issue.path.length - 1] : 'request',
                        message: issue.message,
                    };
                });

                return res.status(400).json({
                    success: false,
                    message: 'Validation Error',
                    errors: formattedErrors,
                });
            }
            next(error);
        }
    };
};