import winston from 'winston';
import DailyRotateFile from 'winston-daily-rotate-file';

const logDir = process.env.LOG_DIR || '/app/logs';

export const logger = winston.createLogger({
  level: 'info',
  format: winston.format.combine(
    winston.format.timestamp(),
    winston.format.json()
  ),
  defaultMeta: { service: 'api-server' },
  transports: [
    new winston.transports.Console(),
    new DailyRotateFile({
      filename: `${logDir}/service-%DATE%.log`,
      datePattern: 'YYYY-MM-DD',
      maxSize: '100m',
      maxFiles: '30d',
      zippedArchive: true,
    }),
  ],
});
