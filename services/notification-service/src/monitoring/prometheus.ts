import { Request, Response, NextFunction } from 'express';
import client from 'prom-client';

client.collectDefaultMetrics({ register: client.register });

export const httpRequestCounter = new client.Counter({
  name: 'http_requests_total',
  help: 'Total number of HTTP requests',
  labelNames: ['method', 'route', 'status_code']
});

export const metricsMiddleware = (req: Request, res: Response, next: NextFunction) => {
  res.on('finish', () => {
    const matchedRoute = req.route ? `${req.baseUrl}${req.route.path}` : req.path;
    
    httpRequestCounter.labels(
      req.method, 
      matchedRoute, 
      res.statusCode.toString()
    ).inc();
  });
  next();
};

export const metricsHandler = async (_req: Request, res: Response) => {
  res.set('Content-Type', client.register.contentType);
  res.send(await client.register.metrics());
};