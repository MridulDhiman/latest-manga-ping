

### Problem statement
CRON job that will run every 24 hrs. It will read the redis, and pick each manga and execute separate goroutines for it. 
In each goroutine, it will search if new manga chapter is present or not. 



