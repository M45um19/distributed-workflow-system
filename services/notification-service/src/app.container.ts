import { NotificationHandler } from './kafka/handlers/notification_handler.js';
import { UserRegisteredHandler } from './kafka/handlers/user_registered.handler.js';
import { KafkaWorker } from './kafka/worker.js';
import { NotificationRepository } from './modules/notification/notification.repository.js';
import { NotificationService } from './modules/notification/notification.service.js';
import { UserRepository } from './modules/user/user.repository.js';
import { UserService } from './modules/user/user.service.js';

export class AppContainer {
  public userService: UserService;
  public notificationService: NotificationService;
  public kafkaWorker?: KafkaWorker;

  constructor(isWorker: boolean) {
    const userRepository = new UserRepository();
    this.userService = new UserService(userRepository);

    const notificationRepository = new NotificationRepository();
    this.notificationService = new NotificationService(notificationRepository);

    if (isWorker) {
      const kWorker = new KafkaWorker();

      const userRegisteredHandler = new UserRegisteredHandler(this.userService);
      kWorker.addTopicHandler('user-registered', userRegisteredHandler);

      const notificationHandler = new NotificationHandler(this.notificationService);
      kWorker.addTopicHandler('send-notification', notificationHandler);

      this.kafkaWorker = kWorker;
      console.info('[Container] Kafka Worker initialized with topic handlers.');
    }
  }
}