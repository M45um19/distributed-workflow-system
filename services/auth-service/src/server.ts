import createApp from "./app";
import { AppContainer } from "./app.container";
import { dbConfig } from "./config/db";
import { env } from "./config/env";
import { grpcConfig } from "./config/grpc";
import { kafkaConfig } from "./config/kafka";
import { redisService } from "./config/redis";
import { registerAuthGrpcService } from "./modules/auth/auth.grpc";

const startServer = async () => {
  try {
    await dbConfig.connect();
    await redisService.ping();
    await kafkaConfig.connect();

    const container = new AppContainer();

    const app = createApp(container);

    registerAuthGrpcService(container.authService);

    grpcConfig.start(50051);

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