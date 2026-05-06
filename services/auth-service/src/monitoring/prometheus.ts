import client from 'prom-client';
import { Request, Response, NextFunction } from 'express';

const collectDefaultMetrics = client.collectDefaultMetrics;
collectDefaultMetrics({ register: client.register });

export const httpRequestCounter = new client.Counter({
  name: 'http_requests_total',
  help: 'Total number of HTTP requests',
  labelNames: ['method', 'route', 'status_code']
});

export const metricsMiddleware = (req: Request, res: Response, next: NextFunction) => {
  res.on('finish', () => {
    httpRequestCounter.labels(
      req.method, 
      req.route ? req.route.path : req.path, 
      res.statusCode.toString()
    ).inc();
  });
  next();
};

export const metricsHandler = async (req: Request, res: Response) => {
  res.set('Content-Type', client.register.contentType);
  res.end(await client.register.metrics());
};