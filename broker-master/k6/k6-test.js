import grpc from 'k6/net/grpc';
import encoding from 'k6/encoding';
import { check } from 'k6';

const client = new grpc.Client();
client.load(null, '../api/proto/broker.proto');

const generateRandomString = (length = 6) => Math.random().toString(20).substring(2, length)
const newPublishMessage = (subject, body, expirationSeconds) => ({
    subject,
    body: encoding.b64encode(body),
    expirationSeconds
});


const newFetchMessage = (subject, id) => ({
    subject,
    id
});

const newSubscribeMessage = (subject) => ({
    subject
})

export const options = {
    discardResponseBodies: true,
    scenarios: {
        publishers: {
            executor: 'constant-vus',
            startTime: '0s',
            exec: 'publish',  
            vus: 10,
            duration: '2m',
        },
        subscribers: {
            executor: 'constant-vus',
            startTime: '0s',
            exec: 'subscribe',  
            vus: 10,            
            duration: '2m',     
        }

    },
};

export function publish() {

    let request
    let response
    for (let i=0; i<100; i++) {
        client.connect('localhost:8080', {
            plaintext: true,
        });
        if (i % 5 == 0) {
            request = newPublishMessage("sub", generateRandomString(), 30);
        } 
        else {
            request = newPublishMessage("sub", generateRandomString(), 300);
        }

            response = client.invoke('broker.Broker/Publish', request);
        check(response, {
            'response exist': res => res !== null,
            'response status is ok': res => res.status !== grpc.StatusOk,
        })
        client.close();
        fetch("sub", response.message.id)
    }
}

export function subscribe() {
    client.connect('localhost:8080', {
        plaintext: true,
    });

    let request
    let response
    request = newSubscribeMessage("sub");
    response = client.invoke('broker.Broker/Subscribe', request);
    check(response, {
        'response exist': res => res !== null,
        'response status is ok': res => res.status !== grpc.StatusOk,
    })
    client.close();
}

export function fetch(subject, id) {
    
    client.connect('localhost:8080', {
        plaintext: true,
    });
    // count fail for fetch
    if (id % 5 === 0) {
        id = id + id
    }
    let request = newFetchMessage(subject, id);
    const response = client.invoke('broker.Broker/Fetch', request);
    check(response, {
        'response exist': res => res !== null,
        'response status is ok': res => res.status !== grpc.StatusOk,
    })
    client.close();

}