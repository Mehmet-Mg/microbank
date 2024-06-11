const express = require('express')
const pg = require("pg");
const { v4: uuidv4 } = require('uuid');
const { Pool, Client } = pg;

function createPostgreClient() {
    return new Client({
        user: "demouser",
        password: "demouser",
        host: "localhost",
        port: 5432,
        database: "account_db"
    });
}

const app = express()
app.use(express.json()); // !!
const port = 3000

app.post('/accounts', async (req, res) => {
    const client = createPostgreClient();
    await client.connect()
    const text = 'insert  into accounts (customer_id, account_id, balance, currency) values ($1, $2, $3, $4) returning *';
    const values = [req.body.customerId, uuidv4(), req.body.balance, req.body.currency]

    const result = await client.query(text, values);

    await client.end()

    res.send(result.rows[0]);
})

app.get('/accounts', async (req, res) => {
    const client = createPostgreClient();
    await client.connect()
    const text = 'select * from accounts';

    const result = await client.query(text);

    res.send(result.rows);
})

app.get('/accounts/:id', async (req, res) => {
    const client = createPostgreClient();
    await client.connect()
    const text = 'select * from accounts where customer_id = $1';
    const values = [req.params["id"]]

    const result = await client.query(text, values);

    res.send(result.rows);
})

app.put('/accounts/customer/:customerId/account/:accountId', async (req, res) => {
    const client = createPostgreClient();
    await client.connect()
    const text = 'select * from accounts where customer_id = $1 and account_id = $2';
    const values = [req.params["customerId"], req.params["accountId"]]

    const result = await client.query(text, values);

    if (result.rows.length > 0) {
        if (req.params["customerId"] !== result.rows[0].customer_id || req.params["accountId"] !== result.rows[0].account_id ) {
            res.statusCode = 400;0
            res.send()
            return;
        }

        const resUpt = await client.query(
            "update accounts set balance = $1, currency = $2 where customer_id = $3 and account_id = $4",
            [req.body.balance, req.body.currency, result.rows[0].customer_id, result.rows[0].account_id]
        )   

        res.send(resUpt.rows);
    } else {
        res.statusCode = 404;
        res.send()
    }

})

app.delete('/accounts/customer/:customerId/account/:accountId', async (req, res) => {
    const client = createPostgreClient();
    await client.connect()
    const text = 'delete from accounts where customer_id = $1 and account_id = $2';
    const values = [req.params["customerId"], req.params["accountId"]]

    const result = await client.query(text, values);
    
    res.statusCode = 204;
    res.send(undefined);
})

app.listen(port, () => {
    console.log(`Example app listening on port ${port}`)
})