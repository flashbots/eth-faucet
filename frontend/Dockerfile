FROM node:20.11.1-slim as build

WORKDIR /app

COPY package.json ./
COPY yarn.lock ./
COPY *.js ./
COPY src ./src
COPY static ./static

RUN apt-get update && apt-get install --yes \
    ca-certificates

RUN yarn install
RUN yarn build

ENV PORT=3000

CMD [ "express.js" ]
