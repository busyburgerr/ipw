FROM node:18.17.1

WORKDIR /app

COPY package*.json .

RUN npm ci

COPY . .

RUN npm run build

CMD [ "node", ".output/server/index.mjs" ]