package main
import ("fmt";"log";"net/http";"os";"github.com/stockyard-dev/stockyard-stampede/internal/server";"github.com/stockyard-dev/stockyard-stampede/internal/store")
func main(){port:=os.Getenv("PORT");if port==""{port="9700"};dataDir:=os.Getenv("DATA_DIR");if dataDir==""{dataDir="./stampede-data"}
db,err:=store.Open(dataDir);if err!=nil{log.Fatalf("stampede: %v",err)};defer db.Close();srv:=server.New(db,server.DefaultLimits())
fmt.Printf("\n  Stampede — Self-hosted load balancer and traffic router\n  Dashboard:  http://localhost:%s/ui\n  API:        http://localhost:%s/api\n  Questions? hello@stockyard.dev — I read every message\n\n",port,port)
log.Printf("stampede: listening on :%s",port);log.Fatal(http.ListenAndServe(":"+port,srv))}
