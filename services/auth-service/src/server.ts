import app from "./app";
import { connectDB } from "./config/db";
import { env } from "./config/env";
import { startGrpcServer } from "./config/grpc";
import redisClient from "./config/redis";
import { registerAuthGrpcService } from "./modules/auth/auth.grpc";

const PORT = env.PORT || 5000;

const startServer = async () => {
  try {

    await connectDB();
    await redisClient.ping();

    registerAuthGrpcService()
    startGrpcServer(50051);

    // Later will add:
    // - Kafka producer init

    app.listen(PORT, () => {
      console.log(`Auth Service running on port: ${PORT}`);
    });

  } catch (error) {
    console.error("Server startup failed:", error);
    process.exit(1);
  }
};

startServer();