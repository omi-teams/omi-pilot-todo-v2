FROM node:24-alpine AS build
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM node:24-alpine
WORKDIR /app
ENV NODE_ENV=production
COPY --from=build /app ./
USER node
EXPOSE 3000
HEALTHCHECK CMD wget -qO- http://localhost:3000/ || exit 1
CMD ["npm", "start"]
