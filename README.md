# STOCKBOT
Bot used to alert stocks in discord using GOLANG and Docker.

# USAGE
Selected alerters can alert stocks, options, or crypto. They can alert specific tickers with specific commands to indicate what they are alerting with their entry price, if one is not provided then it will read the price at time of entry dependent on Polygon API.

The bot will also alert in the alerters channel every ~5% increase in the postion to let people/alerter know that their call is in the green. 

At the end of trading day, the results for the day will be printed to the EOD channel with each call that was green, the highest % it went green, and the author of the alert. 
