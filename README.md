# saga-patterns

This repository describes saga patterns and also introduces an enhancement for choreography pattern which is response control service. You can find technical details in this paper: https://ieeexplore.ieee.org/document/9998226

### Technical background
In saga pattern, there are orchestration and choreography patterns. In orchestration pattern, events are controlled and distributed from a single service, in choreography pattern every service handles its events and triggers other services with a response event without an external controller.

We proposed another solution (extension) for saga pattern which is controlling the responses of microservices that are running under choreography or orchestration pattern using a response control service (RCS).

RCS controls the output of each services. If all services' output are successful, it creates a response for related transaction ID that it is completed successfully totally. It is helpful to understand if a transaction is completed successfully or not under asynchronous environment. Because in asynchronous environment, services are triggered separately and/or parallelly so client does not understand if the request completed successfully or not after triggering the first service. Using RCS, client can be informed when a transaction is completed after some time. 

If there is an error from any of those services, RCS gets this error event and intiates rollbacks for other services, either making the rollback by itself or triggering rollback events for other services.

Using RCS pattern is between classic orchestration and choreography pattern. The main difference between RCS pattern and orchestration pattern is, RCS does not care how inner microservices are communication with each other, only handles the output and initiates the rollback scenario if necessary.

### Demonstration

I used Golang to develop services and RabbitMQ for event controlling. There are 2 microservices `Service A` and `Service B` that runs parallelly. Also there is `RCS` that controls errors of the responses. In this repository, RCS is used only for handling the errors and if there is any error, it makes the rollback.

In `choreography` folder classic choreography pattern is demonstrated. In `choreography_with_response_control` folder, RCS pattern is demonstrated. In each pattern there are Service A and Service B.

Events are distributed under an exchange named `saga_transaction_topic`. Queues (`service_a_queue`, `service_b_queue`, `response_queue`) are binded to this exchange. Service A consumes service_a_queue, Service B consumes service_b_queue, Response Control Service consumes `error_queue`.

### How to run

RabbitMQ is used to handle traffic distribution. Firstly launch RabbitMQ on your local machine:


    docker run -d --hostname my-rabbit --name myrabbit -e RABBITMQ_DEFAULT_USER=admin -e RABBITMQ_DEFAULT_PASS=123456 -p 5672:5672 -p 15672:15672 rabbitmq:3-management

Clone this repository and run queue declarations:


    cd saga-patterns/queueDecleration
	go run .

And run Service A, Service B and RCS



    cd saga-patterns/choreography_with_response_control/serviceA
    go run .
    cd saga-patterns/choreography_with_response_control/serviceB
    go run .
    cd saga-patterns/choreography_with_response_control/response_control_service
    go run .

After that we need to trigger first service which is Service A by publishing message to the exchange with `serviceA` routing key. The payload should be like this:


    {
        "transaction_id": 1,
        "sender": "client"
    }
        
You can publish your message with RabbitMQ UI (http://0.0.0.0:15672/#/exchanges/%2F/saga_transaction_topic)

<img src="https://github.com/mg52/saga-patterns/blob/main/image/rabbitmqui.png" width="600" />

After you publish the first message Service A will handle this and processes its business and it completes successfully then sends an event to Service B. But Service B encounters an error during its process and sends an error event to error queue. RCS gets this error and makes the rollback for Service A which completed its process successfully before.

### Conclusion

There are some benefits and drawbacks of using RCS pattern. 

Benefits,
- It may be easier to track a transaction's final response if it is completed successfully or not by using RCS.
- When RCS makes the rollback on behalf of related service, it reduces the load of the service because related service does not care of the rollback, only makes rollouts.
- If all inner microservices are totally independent from each other, those services can be triggered paralelly and RCS can handle the rollbacks and returning the overall response.

Drawbacks,
- If inner microservices are not totally independent from each other, implementing RCS might be hard and costly. Because RCS needs to make rollback for many scenarios.
- When RCS makes rollbacks on behalf of other services, it needs to be implemented all rollback scenarios for different services into RCS, it might be hard and increases load of RCS if there are too many microservices.
- Making rollbacks on behalf of other services, increases the load of RCS if there are too many microservices.
