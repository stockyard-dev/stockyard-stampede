package main
import ("fmt";"log";"net/http";"os";"github.com/stockyard-dev/stockyard-stampede/internal/server";"github.com/stockyard-dev/stockyard-stampede/internal/store")
func main(){port:=os.Getenv("PORT");if port==""{port="9020"};dataDir:=os.Getenv("DATA_DIR");if dataDir==""{dataDir="./stampede-data"}
db,err:=store.Open(dataDir);if err!=nil{log.Fatalf("stampede: %v",err)};defer db.Close();srv:=server.New(db)
fmt.Printf("\n  Stampede — load balancer and traffic router\n  Dashboard:  http://localhost:%s/ui\n  API:        http://localhost:%s/api\n\n",port,port)
log.Printf("stampede: listening on :%s",port);log.Fatal(http.ListenAndServe(":"+port,srv))}
