This project is to create a to-do app which allows users to create a list of to-do tasks. It is intended for you to use a range of different techniques and features of Golang and tasks followed from top to bottom, building up the application as you go. Don't be afraid to set aside your design at the end of a step if the next step would require massive refactor - it is often simpler to design an application from a clean page and copy in useful code.

A to-do item should include a description and a status of "not started", "started", or "completed". Users should be able to display a list of all to-do items including their statuses, and should be able to update a to-do item to change its status or delete it entirely. Feel free to add any additional functionality if you want.


1) Basic CLI application
- Create a command line application that uses flags to accept a to-do item adds it to an empty list of to-do items and prints the list to console
- After printing the list of to-do items, save them to a file on disk
- When the application starts, load all to-do items from disk before adding new item
- Allow the user to update the description of a to-do item or delete it
2) Advanced functionality
- Use the "log/slog" structured logging package to log errors and when data is saved to disk
- Use the "context" package to add a TraceID to enable traceability of calls through the solution by adding it to all logs
- Separate the core todo store logic into a different package/module to main/CLI code
- Write unit tests to cover usefully testable code
- Use the "os/signal" package and ensure that the application only exits when it receives the interrupt signal (ctrl+c)
3) API
- Use ServeMux in the "net/http" package to expose json http endpoints: "/create", "/get", "/update", and "/delete"
- Add a middleware to create a context with TraceID
4) Web page
- Use "http.FileServer" to serve a static page to a new "/about" endpoint
- Use "html/template" to serve a dynamic page containing a list of all to-do items to a new "/list" endpoint
5) Concurrency
- Use the Actor/Communicating Sequential Processes (CSP) pattern to support concurrent reads and concurrent safe write
- Use Parallel tests to validate that the solution is concurrent safe

Further work:
These tasks are designed to provide you with exercises beyond the end of the the academy course and are more complicated/involved, with less explanation as to how to execute them:

- Repl (Read-eval-print loop)
When the application runs it should ask the user to input text into the console to create, read, update, or delete list items in a loop.

- Multiple startups:
Separate the cli, repl, and api functionality into separate main packages in different modules so that the application can be run as a cli OR a repl OR an api OR all of them together.

- Multi User
The API should include a user ID and support multiple users, each with their own to-do list.

- No interfaces or receivers
Remove all interfaces or receiver functions from your application and instead use the static-singleton pattern.

- Graceful shutdown
When the interrupt signal is sent to the application, it stops accepting incoming http requests and finishes resolving all open http requests before shutting down.

- Publish modules
Ensure the core todo store logic is in a self contain module and publish V1 to github, then use the published module in the application entry points rather than local one.

- Benchmark
Use benchmark unit tests to determine the performance of your application. Run a separate application to bombard your code with requests and evaluate the performance.

- pprof
Use the pprof utility to profile application performance.

- Sharding
Split the todo store "back end" module into a separate executable to the "front end" api/cli/repl/web server, run multiple instances of the back end and distribute traffic from the front end using ring hashing.

- GRPC
Enable communication between front end and back end with GRPC.

## Links
[https://gobyexample.com/](https://gobyexample.com/)  
[https://www.oreilly.com/library/view/learning-go-2nd/9781098139285/](https://www.oreilly.com/library/view/learning-go-2nd/9781098139285/)  
[https://youtu.be/YzLrWHZa-Kc?si=F-UDjZc6eARc1Uws](https://youtu.be/YzLrWHZa-Kc?si=F-UDjZc6eARc1Uws)  
[https://quii.gitbook.io/learn-go-with-tests](https://quii.gitbook.io/learn-go-with-tests) (although this teaches a much more OO style)