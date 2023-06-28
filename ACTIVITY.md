# 🧑‍💻 Hands-on Activity

## 💔 Problem

The users are reporting that using TODO lists has become slow sometimes. They are upset, and customer service are being
overwhelmed with calls, emails etc.

Your frontend developers insist that the problem is not in their Node.js app.

You've recently heard about Distributed Tracing, and think it could help identify what is happening. The frontend app
is already using it.

## 🥅 Goals

1. Add and configure tracing for the application (with OpenTelemetry); then
2. Use traces to identify what makes certain requests slow

> 🧘 **Focus**
> A lot of the code here is to support the CRUD app, and not a focus for this session.
>
> For the interactive session, we'll be focusing on only:
>
> 1. Adding Tracing to the app
> 2. Identifying the issue(s) from the traces
>
> For this, we'll only be looking at 2 files:
>
> * [backend/cmd/root.go](backend/cmd/root.go) - application setup
> * [backend/todo/routes/router.go](backend/todo/routes/router.go) - HTTP route setup

## 🤷 Assumptions

It is assumed that the pre-requisites from the `README.md` have been completed before starting this activity.

## 🔄 Steps

> 🤔 **As you go**
>
> When you encounter a 🗣**Discuss** note in the steps to follow, have a think about the question.
> Discuss with your table when others are ready, either as you go, or when you've completed the steps.

### 🔧 Part 1: A running app

> 🙋**Help!**
>
> If you run into problems at any stage here, let someone at your table or in the room know, and we'll try to help!

1. [ ] 🖥 Open a shell terminal and navigate to the root of the cloned repository.

2. [ ] 🔁 Start the services the app depends on:

    ```shell
    docker compose up -d
    ```

3. [ ] ⬆️ Check that everything has started successfully.

    ```shell
    docker compose ps
    ```

   You should see something like below, with `running` for both the database and jaeger.

    ```text
    NAME                 COMMAND                  SERVICE             STATUS              PORTS
    todo-list-db-1       "docker-entrypoint.s…"   db                  running             127.0.0.1:5432->5432/tcp
    todo-list-jaeger-1   "/go/bin/all-in-one-…"   jaeger              running             5775/udp, 5778/tcp, 127.0.0.1:14268->14268/tcp, 6831-6832/udp, 14250/tcp, 127.0.0.1:16686->16686/tcp
    ```

4. [ ] 🌏 Open [Jaeger UI](http://localhost:16686) in your web browser. Keep this open, you'll need it later

5. [ ] 🧭 In your terminal, change to the `backend` directory (where the Go module is located):

    ```shell
    cd backend
    ```

6. [ ] 🧑‍💻 Compile and run the backend app

    ```shell
    go build . && ./backend
    ```

   If it has started successfully, you should see:

    ```text
    Ready to accept requests on http://127.0.0.1:8080
    ```

7. [ ] 🖥 Open a second shell terminal at the root of the repo, and simulate API traffic:

    ```shell
    docker compose --profile frontend up simulate-ui
    ```

8. [ ] 👀 Review the output from the backend app (other terminal).

   You should see something like:

    ```text
    Ready to accept requests on http://127.0.0.1:8080
    {"time":"2023-06-22T17:10:17.75235+10:00","level":"INFO","msg":"HTTP Request Success","http.path":"/api/v1/lists/235","http.route":"/api/v1/lists/:list_id","http.status":200,"http.request_duration":"34.984333ms"}
    {"time":"2023-06-22T17:10:17.770517+10:00","level":"INFO","msg":"HTTP Request Success","http.path":"/api/v1/lists/235/items","http.route":"/api/v1/lists/:list_id/items","http.status":200,"http.request_duration":"4.55875ms"}
    {"time":"2023-06-22T17:10:17.790419+10:00","level":"INFO","msg":"HTTP Request Success","http.path":"/api/v1/lists/447","http.route":"/api/v1/lists/:list_id","http.status":200,"http.request_duration":"5.983459ms"}
    ...
    ```

   > 🗣**Discuss**
   >
   > What can we identify about the requests from just the logs?

   > 🧙 **Tip**
   >
   > You could review traces here, but they won't include our app since it isn't instrumented yet.

9. [ ] 💀 Shutdown the backend before moving on - use `Control` + `C`

### 🔧 Part 2: Adding Tracing

Now that we've confirmed everything is working, we can add in tracing

1. [ ] 📝 Open [backend/cmd/root.go](backend/cmd/root.go), and edit the `Run` function:

    ```go
    func Run(ctx context.Context, cfg *Config, logger *slog.Logger, stdout, stderr io.Writer) error {
        ...
    
        // Setup Tracing: Uncomment this block
        //shutdown, err := setupTracing(ctx, logger)
        //if err != nil {
        //	return err
        //}
        //
        //defer shutdown()
        
        ...
    ```

   You should find the block above. Uncomment the code to enable the tracing provider.

   > 🗣**Discuss**
   >
   > What is happening in the `setupTracing` function?

2. [ ] 📝 Open [backend/todo/routes/router.go](backend/todo/routes/router.go), and edit the `handler` function:

    ```go
    func (m *mux) handler(method, route string, h http.Handler) {
        w := requestLog(h, m.logger, route)
        // Instrument HTTP Handlers: Uncomment the line below
        //w = otelhttp.NewHandler(w, method+" "+route)
        
        ...
    ```

   You should find the block above. Uncomment the line containing `otelhttp.NewHandler`

   > 🗣**Discuss**
   >
   > * How does this impact the handling of HTTP requests?
   > * Why does this need to be applied here?

3. [ ] 🧑‍💻 Compile and run the backend app again, and simulate traffic

   🖥 Terminal 1 (dir: `{repo}/backend`):
    ```shell
    go build . && ./backend
    ```

   🖥 Terminal 2 (dir: `{repo}`):
    ```shell
    docker compose --profile frontend up simulate-ui
    ```

4. [ ] 🌏 Switch back to [Jaeger UI](http://localhost:16686) which you opened earlier, and refresh the page

5. [ ] 🕵️ Select `todo-list-api` from the _Services_ drop-down, and click _Find Traces_

   > 🙋 **Missing Service?**
   >
   > Don't see `todo-list-api` in the list of Services? Trace data may not have been sent - check app logs

6. [ ] 📊 Select and review the traces for the "User Homepage" and "TODO List" UI requests

   > 🧙 **Tip**
   >
   > There are details that expand when clicked on, e.g. the visual span, it's "Tags" and "Process"

   > 🗣**Discuss**
   >
   > * How does this compare to what we found in the logs?
   > * Is there enough detail here to identify the issue?

### 🔧 Part 3: More Instrumentation

> 📈 **Progress**
>
> The instrumentation so far was helpful to see how the requests fit together, identify slow requests, and some extra
> details about them.
>
> **However**, we're still in the dark about why our app is taking so long to handling these requests...

1. [ ] 📝 Open [backend/cmd/root.go](backend/cmd/root.go) again, and edit the `openDB` function:

    ```go
    func openDB(connURL string) (*sql.DB, error) {
        ...
    
        return sql.OpenDB(conn), nil
        // Instrument Database Calls: Replace the line above with the one below
        //return traceDB(connURL, conn)
    ```

   You should find the block above. Replace the existing `return` with the commented one

2. [ ] 🧑‍💻 Compile and run the backend app again, and simulate traffic

   🖥 Terminal 1 (dir: `{repo}/backend`):
    ```shell
    go build . && ./backend
    ```

   🖥 Terminal 2 (dir: `{repo}`):
    ```shell
    docker compose --profile frontend up simulate-ui
    ```

3. [ ] 🌏 Switch back to [Jaeger UI](http://localhost:16686), and review the new traces

   > 🧙 **Tip**
   >
   > You can expand and collapse child spans in the 'Service & Operation' section

   > 🗣 **Discuss**
   >
   > Why are these requests taking so long?

## 👀 Review

With your table group, refer to any "🗣**Discuss**" points you've yet to cover, and compare/share thoughts.

You can also discuss anything else related to the activity, your findings etc.

## 🌏 Resources

* https://opentelemetry.io/docs/instrumentation/go/getting-started/
* https://github.com/open-telemetry/opentelemetry-go-contrib
* https://github.com/XSAM/otelsql
