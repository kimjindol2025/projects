# Node.js 기반 블로그 서버
FROM node:20-alpine

WORKDIR /app

# 필수 파일 복사
COPY package.json .
COPY server-comments.js .
COPY posts/ ./posts/
COPY index.html .

# 포트 노출
EXPOSE 8253

# 서버 시작
CMD ["node", "server-comments.js", ":8253"]
