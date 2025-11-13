const axios = require('axios');
const amqp = require('amqplib');

const CANCHAS_URL = process.env.CANCHAS_API_URL || 'http://localhost:8081';
const RABBIT_URL = process.env.RABBITMQ_URL || 'amqp://guest:guest@localhost:5672/';
const EXCHANGE = process.env.RABBITMQ_EXCHANGE || 'canchas_events';

async function main() {
    try {
        console.log(`Fetching canchas from ${CANCHAS_URL}/canchas`);
        const res = await axios.get(`${CANCHAS_URL}/canchas`);
        const canchas = res.data && res.data.canchas ? res.data.canchas : res.data;
        if (!Array.isArray(canchas) || canchas.length === 0) {
            console.log('No canchas found to reindex.');
            return;
        }

        console.log(`Found ${canchas.length} canchas. Connecting to RabbitMQ ${RABBIT_URL}...`);
        const conn = await amqp.connect(RABBIT_URL);
        const ch = await conn.createChannel();

        await ch.assertExchange(EXCHANGE, 'topic', { durable: true });

        for (const cancha of canchas) {
            const event = {
                type: 'create',
                entity: 'cancha',
                entity_id: cancha.id || cancha.ID || cancha._id || cancha.id,
            };

            const routingKey = 'cancha.create';
            ch.publish(EXCHANGE, routingKey, Buffer.from(JSON.stringify(event)), {
                persistent: true,
                contentType: 'application/json',
            });

            console.log(`Published event for cancha id=${event.entity_id} number=${cancha.number || 'n/a'}`);
        }

        await ch.close();
        await conn.close();
        console.log('Reindexing events published successfully.');
    } catch (err) {
        console.error('Error during reindex:', err.message || err);
        process.exitCode = 2;
    }
}

main();
