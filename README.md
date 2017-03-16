Todo
-Genreal broadcast function that sends from 1 channel to 1-n channels
-Rydde opp koden
  
Log on another computer"ssh username@ip"
    ex: "ssh student@129.241.187.144"
Transfer file: "scp -r pathSource username@ip:pathTarget"
    eks: "scp -r ~/Documents/Gruppe38/project/main student@129.241.187.144:~/Documents/gruppe38"
    
    Endringer etter levering
-IP adresse til labplass 11, 10 lagt til i definitions.go
- Changed copy and copy
-in AssignMovmentInstrucions in ElevatorManagement add initial status to AtFloor
- changed condition to move from deadElevator and from noNetwork to also demanding no more active orders
