import createApp from "./app";
import { AppContainer } from "./app.container";
import { connectDB } from "./config/db";
import { env } from "./config/env";
import { startGrpcServer } from "./config/grpc";
import redisClient from "./config/redis";
import { registerAuthGrpcService } from "./modules/auth/auth.grpc";

const startServer = async () => {
  try {
    await connectDB();
    await redisClient.ping();

    const container = new AppContainer();

    const app = createApp(container);

    registerAuthGrpcService(container.authService);
    startGrpcServer(50051);

    const PORT = env.PORT || 5000;
    app.listen(PORT, () => {
      console.log(`Auth Service running on port: ${PORT}`);
    });

  } catch (error) {
    console.error("Server startup failed:", error);
    process.exit(1);
  }
};

startServer();