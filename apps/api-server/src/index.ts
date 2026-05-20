import app from './app';
import { config } from './config';
import { logger } from './logger';

const PORT = config.PORT;

app.listen(PORT, () => {
  logger.info('API server running', { port: PORT });
});
