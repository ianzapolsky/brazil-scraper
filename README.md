# Instructions to run

1. Download Golang
2. In this directory, run:
   ```bash
   mkdir DATUMS
   ```
3. From the command line, run
   ```golang
   go run scraper.go > log.txt
   ```
4. Wait...
5. In a separate terminal pane, monitor progress by running the following:
   ```bash
   cd DATUMS
   watch 'ls | wc -l'
   ```
