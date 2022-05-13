FROM node:16.10.0

WORKDIR /srv

COPY ./package.json ./
COPY ./package-lock.json ./
RUN npm install

COPY . ./
RUN npm run build

CMD ["npm", "start"]
