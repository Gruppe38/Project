..# Project.

set GOPATH=C:\Users\Arild\Desktop\Lab\Project

Simulator: fra min laptop
source ~/dlang/dmd-2.073.0/activate
cd Documents/git/Simulator/Project/simulator/server
rdmd sim_server.d

Undersøke hva som skjer om vi er under 1,  over 4. Passe på at den aldri kjører dit selv. 

Idé: endre messageID til å starte på 0 og telle oppover, delt for alle typer meldinger, men hver pc har sin egen teller. Eneste endring på mottaker side er at vi må ikke bare huske hvilke meldinger vi har sendt, men også hvem vi sendte dem til.
  Mulighet som da åpner seg: Bare videresende ordermessage om den har høyere id (er nyere) en sist mottate. Dermed er vi sikret at en gammel ordermessage ikke overlagrer den gamle. 
  
  
  ip for labplass 14: 129.241.187.142
  ip for labplass 14: 129.241.187.148
  ip for labplass 13: 129.241.187.152
  
  freezelog:
  1: localElevator() got new movinst {false false 2}
  2: New button: DOWN#
  3 recieved statusReport in createORderQueue(): from elevator 1
  4 watchIncommindOrders() did not send button CDM2
