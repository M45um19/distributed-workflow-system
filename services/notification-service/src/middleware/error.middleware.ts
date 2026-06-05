import { NextFunction, Request, Response } from "express";

import { env } from "../config/env.js";
import { AppError } from "../utils/appError.js";

interface MongoCastError extends Error {
  path: string;
  value: string;
}

const sendErrorDev = (err: AppError, res: Response) => {
  res.status(err.statusCode || 500).json({
    success: false,
    message: err.message,
    stack: err.stack,
    error: err,
  });
};

const sendErrorProd = (err: AppError, res: Response) => {
  if (err.isOperational) {
    res.status(err.statusCode || 500).json({
      success: false,
      message: err.message,
    });
  } else {
    console.error("ERROR : ", err);
    res.status(500).json({
      success: false,
      message: "Something went very wrong!",
    });
  }
};

export const globalErrorHandler = (
  err: Error | AppError,
  _req: Request,
  res: Response,
  _next: NextFunction
) => {
  let statusCode = 500;
  if (err instanceof AppError) {
    statusCode = err.statusCode;
  }

  if (env.NODE_ENV === "development") {

    const devError = err instanceof AppError ? err : new AppError(err.message, statusCode);
    sendErrorDev(devError, res);
  } else {
    let error: AppError;

    if (err.name === "CastError") {
      const castErr = err as MongoCastError;
      error = new AppError(`Invalid ${castErr.path}: ${castErr.value}`, 400);
    } else if (err.name === "JsonWebTokenError") {
      error = new AppError("Invalid token. Please log in again!", 401);
    } else if (err.name === "KafkaJSProtocolError") {
      error = new AppError("Message queue is temporarily unavailable", 503);
    } else if (err.name === "KafkaJSConnectionError") {
      error = new AppError("Could not connect to the event broker", 503);
    } else if (err instanceof AppError) {
      error = err;
    } else {
      error = new AppError(err.message, 500);
    }

    sendErrorProd(error, res);
  }
};