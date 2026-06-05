import { UserRegisteredHandler } from './kafka/handlers/user_registered.handler.js';
import { KafkaWorker } from './kafka/worker.js';
import { UserRepository } from './modules/user/user.repository.js';
import { UserService } from './modules/user/user.service.js';

export class AppContainer {
  public userService: UserService;
  public kafkaWorker?: KafkaWorker;

  constructor(isWorker: boolean) {
    const userRepository = new UserRepository();
    this.userService = new UserService(userRepository);

    if (isWorker) {
      const kWorker = new KafkaWorker();

      const userRegisteredHandler = new UserRegisteredHandler(this.userService);
      kWorker.addTopicHandler('user-registered', userRegisteredHandler);

      this.kafkaWorker = kWorker;
      console.info('[Container] Kafka Worker initialized with topic handlers.');
    }
  }
}