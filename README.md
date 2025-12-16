# Golang Developer Assignment

Develop in Go language a service that will provide an API for retrieval of the Last Traded Price of Bitcoin for the following currency pairs:

1. BTC/USD
2. BTC/CHF
3. BTC/EUR

The request path is:
`/api/v1/ltp`

The response shall constitute JSON of the following structure:
```json
{
  "ltp": [
    {
      "pair": "BTC/CHF",
      "amount": 49000.12
    },
    {
      "pair": "BTC/EUR",
      "amount": 50000.12
    },
    {
      "pair": "BTC/USD",
      "amount": 52000.12
    }
  ]
}
```

## Requirements:

1. The incoming request can done for as for a single pair as well for a list of them
2. You shall provide time accuracy of the data up to the last minute.
3. Code shall be hosted in a remote public repository
4. readme.md includes clear steps to build and run the app
5. Integration tests
6. Dockerized application

## Docs
The public Kraken API might be used to retrieve the above LTP information
[API Documentation](https://docs.kraken.com/rest/#tag/Spot-Market-Data/operation/getTickerInformation)
(The values of the last traded price is called "last trade closed")

