# Signature Service - Coding Challenge

Thank you for a fascinating task! I tried to comment my code, but I also want to make some statements here:

- I upped the version of Go to 1.22 in order to use the new http std package pattern matching.
- I used testify for testing, because that's what I would prefer to use in my newer projects, though I'm familiar with
  the standard table testing.
- I tested the domain logic thoroughly, but I didn't test the http handlers and other services. Usually, I would test
  them as well, but I wanted to keep the code concise so you would have time to review all of that :)
- In-memory database is used, so I decided to put the mutex in each device so that each signature operation would be
  atomic.
- I provide the context to the persistence layer, but I didn't use it. I would use it in a real project to cancel the
  operation if the context is done.
- The app can be configured with both env vars and .env file. I used my own library for that, but I'm also familiar with
  viper.
- I decided to use zap for logging as I have more experience with it.
- I've put a small .http file in `test/http` to make test calls to the api.

### How to run

Please, use the `Makefile` to run the app. Given more time I would also create a docker-compose file to run the app.

**In case of other questions, please let me know. I'm looking forward to your feedback!**