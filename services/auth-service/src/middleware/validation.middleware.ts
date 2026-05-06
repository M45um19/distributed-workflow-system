import { Request, Response, NextFunction } from 'express';
import { ZodError, ZodSchema } from 'zod';

export const validateRequest = <T>(schema: ZodSchema<T>) => {
    return async (req: Request, res: Response, next: NextFunction) => {
        try {
            const validatedData = await schema.parseAsync({
                body: req.body,
                query: req.query,
                params: req.params,
            });

            const data = validatedData as { 
                body: typeof req.body; 
                query: typeof req.query; 
                params: typeof req.params 
            };

            req.body = data.body;
            req.query = data.query;
            req.params = data.params as Record<string, string>;

            next();
        } catch (error: unknown) {
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