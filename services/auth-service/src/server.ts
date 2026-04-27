import app from "./app";
import { connectDB } from "./config/db";
import { env } from "./config/env";
import redisClient from "./config/redis";

const PORT = env.PORT || 5000;

const startServer = async () => {
  try {

    await connectDB();
    await redisClient.ping();
    // Later will add:
    // - Redis connection
    // - Kafka producer init
    // - gRPC server start

    app.listen(PORT, () => {
      console.log(`Auth Service running on port ${PORT}`);
    });

  } catch (error) {
    console.error("Server startup failed:", error);
    process.exit(1);
  }
};

startServer();