import mongoose from "mongoose";

import { env } from "./env.js";

export interface IDatabase {
  connect(): Promise<void>;
  disconnect(): Promise<void>;
}

class DatabaseConfig implements IDatabase {
  constructor() {
    this.setupEventListeners();
  }

  private setupEventListeners(): void {
    mongoose.connection.on("connected", () => {
      console.info("MongoDB connected successfully.");
    });

    mongoose.connection.on("error", (err) => {
      console.error(`MongoDB connection error: ${err}`);
    });

    mongoose.connection.on("disconnected", () => {
      console.warn("MongoDB connection disconnected. Attempting to reconnect...");
    });
  }

  public async connect(): Promise<void> {
    if (mongoose.connection.readyState === 1 || mongoose.connection.readyState === 2) {
      return;
    }

    const MONGO_URI = env.MONGO_URI;

    if (!MONGO_URI) {
      throw new Error("MONGO_URI is not defined in environment variables.");
    }

    try {
      await mongoose.connect(MONGO_URI, {
        autoIndex: true,
        serverSelectionTimeoutMS: 5000,
      });
    } catch (error) {
      console.error("Critical Error: MongoDB connection could not be established.");
      throw error; 
    }
  }

  public async disconnect(): Promise<void> {
    if (mongoose.connection.readyState === 0) {
      return;
    }

    try {
      await mongoose.disconnect();
      console.info("🔌 MongoDB connection closed safely.");
    } catch (error) {
      console.error("Error during MongoDB disconnection:", error);
    }
  }
}

export const dbConfig = new DatabaseConfig();