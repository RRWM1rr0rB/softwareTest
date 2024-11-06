1. your task is executed in the form of a microservice.
2. code with gRPC contracts - https://github.com/WM1rr0rB8/contractsTest/tree/main
3. Code with personal libraries that I use in my projects - https://github.com/WM1rr0rB8/librariesTest/tree/main/backend/golang , which I used to write the microservice .
4. In the app/test folder there are test requests for HTTP and gRPC.
5. Apart from one method, I have written a method to change the status in the Order.
6. Also a method for searching orders, I added sorting, pagination and search by fields and search itself.
7. In the microservice, I also added graphane and prometheus metrics, tracing and logging.
8. The microservice is divided into layers:
   controller - interactions with other services or external handle, data mapping, filters, and method handling.
9. policy - business logic, error handling.
10. domain - includes service and storage.
11. The size is put in the config to allow for dynamic changes.