import { AppContainer } from "./app.container.js";
import createApp from "./app.js";
import { dbConfig } from "./config/db.js";
import { env } from "./config/env.js";
import { grpcConfig } from "./config/grpc.js";
import { kafkaConfig } from "./config/kafka.js";
import { redisService } from "./config/redis.js";
import { registerAuthGrpcService } from "./modules/auth/auth.grpc.js";

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