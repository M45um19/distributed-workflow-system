import { diag, DiagConsoleLogger, DiagLogLevel } from '@opentelemetry/api';
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-grpc';
import { GrpcInstrumentation } from '@opentelemetry/instrumentation-grpc';
import { HttpInstrumentation } from '@opentelemetry/instrumentation-http';
import { Resource } from '@opentelemetry/resources';
import { NodeSDK } from '@opentelemetry/sdk-node';
import { SemanticResourceAttributes } from '@opentelemetry/semantic-conventions';

import { env } from '../config/env.js';

// Enable OpenTelemetry SDK internal diagnostic logging
diag.setLogger(new DiagConsoleLogger(), DiagLogLevel.INFO);

const resource = new Resource({
  [SemanticResourceAttributes.SERVICE_NAME]: env.SERVICE_NAME,
});

const traceExporter = new OTLPTraceExporter({
  url: env.OTEL_EXPORTER_OTLP_ENDPOINT as string, 
});

const sdk = new NodeSDK({
  resource: resource,
  traceExporter: traceExporter,
  instrumentations: [
    new GrpcInstrumentation(),
    new HttpInstrumentation(),
  ],
});

try {
  sdk.start();
  console.log('OpenTelemetry SDK initialized successfully.');
} catch (error) {
  console.error('Error initializing OpenTelemetry SDK:', error);
}

process.on('SIGTERM', async () => {
  console.log('Shutting down OpenTelemetry SDK...');
  await sdk.shutdown()
    .then(() => console.log('Tracing terminated cleanly.'))
    .catch((error) => console.error('Error during tracing shutdown:', error));
});

export default sdk;