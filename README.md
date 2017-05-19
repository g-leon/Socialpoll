## System Design

- `twittervotes` pulls the relevant tweet data via the Twitter API
, decides what is being voted for (rather, which options are mentioned in the body), 
and then pushes the vote into NSQ
- `counter` listens out for votes on the messaging queue and
periodically saves results in the MongoDB database. It receives
the vote messages from NSQ and keeps an in-memory counter of the
results, periodically pushing it to persist the data.
- `web` is a web server program that will expose the live results.

## Quick setup
1. In the top-level folder, start the `nsqlookup` daemon:
    
        nsqlookup
    
2. In the same directory, start the `nsqd` daemon:
        
        nsqd --lookup-tcp-address=localhost:4160
         
3. Start the MongoDB daemon:
        
        mongod
        
4. Navigate to the `counter` folder and build and run it:

        cd counter
        go build -o counter
        ./counter
        
5. Navigate to the `twittervotes` folder and build and runt it. Ensure that you have the appropriate environment variables set;
otherwise, you will see errors when you run the program:
    
        cd ../twittervotes
        go build -o twittervotes
        ./twittervotes
        
6. Navigate to the `api` folder and build and run it:

        cd ../api
        go build -o api
        ./api
             
7. Navigate to the `web` folder and build and run it:
        
        cd ../web
        go build -o web
        ./web
        
8. Open a browser and head to `http://localhost:8081/`. Using the
user interface, create a poll called "Moods" and input some common enough words 
as options, such as "happy", "sad", "fail", "success".
Once you have created the poll, you will be taken to the view page
where you will start to see the results coming in. Wait for 
a few seconds and see UI updates in real time, showing live, real-time results.