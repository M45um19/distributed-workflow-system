import { NextFunction, Request, Response } from "express";
import { AppError } from "../utils/appError";
import { env } from "../config/env";

const sendErrorDev = (err: any, res: Response) => {
  res.status(err.statusCode).json({
    success: false,
    message: err.message,
    stack: err.stack,
    error: err,
  });
};

const sendErrorProd = (err: any, res: Response) => {
  if (err.isOperational) {
    res.status(err.statusCode).json({
      success: false,
      message: err.message,
    });
  } 
  else {
    console.error("ERROR: ", err);
    res.status(500).json({
      success: false,
      message: "Something went very wrong!",
    });
  }
};

export const globalErrorHandler = (
  err: any,
  req: Request,
  res: Response,
  next: NextFunction,
) => {
  err.statusCode = err.statusCode || 500;
  err.status = err.status || "error";

  if (env.NODE_ENV === "development") {
    sendErrorDev(err, res);
  } else {
    let error = { ...err };
    error.message = err.message;

    if (err.name === "CastError") error = new AppError(`Invalid ${err.path}: ${err.value}`, 400);
    if (err.name === "JsonWebTokenError") error = new AppError("Invalid token. Please log in again!", 401);

    sendErrorProd(error, res);
  }
};