
FROM node:20-alpine AS builder

WORKDIR /app

COPY services/notification-service/package*.json ./

RUN npm ci

COPY services/notification-service/ .

RUN npm run build

RUN npm prune --production



FROM node:20-alpine

WORKDIR /app

RUN mkdir -p /shared-proto
COPY ./shared-proto /shared-proto

COPY --from=builder /app/package*.json ./
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/dist ./dist

RUN chown -R node:node /app /shared-proto
USER node

EXPOSE 5001

CMD ["node", "dist/server.js"]