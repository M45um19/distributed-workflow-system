import { Response } from "express";

export interface IApiMeta {
  page?: number;
  limit?: number;
  total?: number;
  totalPage?: number;
}

export interface IApiResponse<T> {
  statusCode: number;
  success: boolean;
  message: string;
  data: T;
  meta?: IApiMeta;
}

export const sendResponse = <T>(res: Response, data: IApiResponse<T>): void => {
  const responseData: IApiResponse<T> = {
    statusCode: data.statusCode,
    success: data.success,
    message: data.message,
    data: data.data,
  };

  if (data.meta) {
    responseData.meta = data.meta;
  }

  res.status(data.statusCode).json(responseData);
};