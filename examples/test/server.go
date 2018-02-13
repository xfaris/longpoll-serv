package main

import (
	"log"
	"net/http"

	"github.com/jcuga/golongpoll"
	"fmt"
)

func main() {
	manager, err := golongpoll.StartLongpoll(golongpoll.Options{
		LoggingEnabled:                 true,
		MaxLongpollTimeoutSeconds:      120,
		MaxEventBufferSize:             100,
		EventTimeToLiveSeconds:         60 * 2, // Event's stick around for 2 minutes
		DeleteEventAfterFirstRetrieval: false,
	})
	if err != nil {
		log.Fatalf("Failed to create manager: %q", err)
	}

	http.HandleFunc("/test", ExampleHomepage)
	// Serve handler that generates events
	http.HandleFunc("/test/action", getUserActionHandler(manager))
	// Serve handler that subscribes to events.
	http.HandleFunc("/test/events", manager.SubscriptionHandler)
	// Start webserver
	log.Fatal(http.ListenAndServe("127.0.0.1:8081", nil))

	manager.Shutdown() // Stops the internal goroutine that provides subscription behavior
}

// A fairly trivial json-convertable structure that demonstrates how events
// don't have to be a plain string.  Anything JSON will work.
type RoomAction struct {
	Action   string `json:"action"`
}

// Creates a closure function that is used as an http handler that allows
// users to publish events (what this example is calling a user action event)
func getUserActionHandler(manager *golongpoll.LongpollManager) func(w http.ResponseWriter, r *http.Request) {
	// Creates closure that captures the LongpollManager
	return func(w http.ResponseWriter, r *http.Request) {
		rid := r.URL.Query().Get("rid")
		action := r.URL.Query().Get("action")
		// Perform validation on url query params:
		if len(rid) == 0 || len(action) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Missing required URL param."))
			return
		}
		actionEvent := RoomAction{Action: action}
		manager.Publish(rid, actionEvent)
	}
}

func ExampleHomepage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `
<html>
<head>
    <title>golongpoll basic example</title>
</head>
<body>
    <h1>golongpoll basic example</h1>
    <h2>Here's whats happening around the farm:</h2>
    <ul id="animal-events"></ul>
<script src="http://code.jquery.com/jquery-1.11.3.min.js"></script>
<script>

    // for browsers that don't have console
    if(typeof window.console == 'undefined') { window.console = {log: function (msg) {} }; }

    // Start checking for any events that occurred after page load time (right now)
    // Notice how we use .getTime() to have num milliseconds since epoch in UTC
    // This is the time format the longpoll server uses.
    var sinceTime = (new Date(Date.now())).getTime();

    // Let's subscribe to animal related events.
    var category = "farm";

    (function poll() {
        var timeout = 45;  // in seconds
        var optionalSince = "";
        if (sinceTime) {
            optionalSince = "&since_time=" + sinceTime;
        }
        var pollUrl = "/test/events?timeout=" + timeout + "&category=" + category + optionalSince;
        // how long to wait before starting next longpoll request in each case:
        var successDelay = 10;  // 10 ms
        var errorDelay = 3000;  // 3 sec
        $.ajax({ url: pollUrl,
            success: function(data) {
                if (data && data.events && data.events.length > 0) {
                    // got events, process them
                    // NOTE: these events are in chronological order (oldest first)
                    for (var i = 0; i < data.events.length; i++) {
                        // Display event
                        var event = data.events[i];
                        $("#animal-events").append("<li>" + event.data + " at " + (new Date(event.timestamp).toLocaleTimeString()) +  "</li>")
                        // Update sinceTime to only request events that occurred after this one.
                        sinceTime = event.timestamp;
                    }
                    // success!  start next longpoll
                    setTimeout(poll, successDelay);
                    return;
                }
                if (data && data.timeout) {
                    console.log("No events, checking again.");
                    // no events within timeout window, start another longpoll:
                    setTimeout(poll, successDelay);
                    return;
                }
                if (data && data.error) {
                    console.log("Error response: " + data.error);
                    console.log("Trying again shortly...")
                    setTimeout(poll, errorDelay);
                    return;
                }
                // We should have gotten one of the above 3 cases:
                // either nonempty event data, a timeout, or an error.
                console.log("Didn't get expected event data, try again shortly...");
                setTimeout(poll, errorDelay);
            }, dataType: "json",
        error: function (data) {
            console.log("Error in ajax request--trying again shortly...");
            setTimeout(poll, errorDelay);  // 3s
        }
        });
    })();
</script>
</body>
</html>`)
}