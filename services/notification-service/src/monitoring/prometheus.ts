import { Request, Response, NextFunction } from 'express';
import client from 'prom-client';

client.collectDefaultMetrics({ register: client.register });

export const httpRequestCounter = new client.Counter({
  name: 'http_requests_total',
  help: 'Total number of HTTP requests',
  labelNames: ['method', 'route', 'status_code']
});

const getRoutePattern = (req: Request, res: Response): string => {
  if (req.route) {
    return `${req.baseUrl}${req.route.path}`;
  }

  // Prevent cardinality explosion on 404 / unmatched paths
  if (res.statusCode === 404) {
    return 'NOT_FOUND';
  }

  // Normalize fallback path (e.g. middleware rejections)
  return req.path
    .replace(/[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}/g, ':id')
    .replace(/\b[0-9a-fA-F]{24}\b/g, ':id')
    .replace(/\/\d+(?=\/|$)/g, '/:id');
};

export const metricsMiddleware = (req: Request, res: Response, next: NextFunction) => {
  res.on('finish', () => {
    const matchedRoute = getRoutePattern(req, res);
    
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