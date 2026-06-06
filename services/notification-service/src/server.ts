import http from 'http';

import { AppContainer } from './app.container.js';
import createApp from './app.js';
import { dbConfig } from './config/db.js';
import { env } from './config/env.js';
import { redisService } from './config/redis.js';
import { socketConfig } from './config/socket.js';

const startServer = async () => {
  try {
    await dbConfig.connect();
    await redisService.ping();

    const container = new AppContainer(false);
    const app = createApp(container);

    const server = http.createServer(app);
    socketConfig.init(server);
    const PORT = env.PORT;
    server.listen(PORT, () => {
      console.log(`Notification API & Socket Server running on port: ${PORT}`);
    });

  } catch (error) {
    console.error('Notification Server startup failed:', error);
    process.exit(1);
  }
};

startServer();