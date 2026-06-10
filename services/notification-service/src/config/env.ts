import dotenv from "dotenv";

dotenv.config();

export const env = {
  NODE_ENV: process.env.NODE_ENV,
  PORT: process.env.PORT,
  MONGO_URI: process.env.MONGO_URI,
  REDIS_URI: process.env.REDIS_URI,
  KAFKA_BROKERS: process.env.KAFKA_BROKERS,
  AUTH_SERVICE_GRPC_ADDRESS: process.env.AUTH_SERVICE_GRPC_ADDRESS,
  JWT_SECRET: process.env.JWT_SECRET,
  AUTH_PROTO_PATH: process.env.AUTH_PROTO_PATH
};