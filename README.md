# Progetto SDCC – Service Registry con gRPC

Progetto svolto individualmente.


## Gestione Sessione

- Lookup eseguito solo all’avvio  
- La lista server viene riusata  
- È presente comando “Ricarica lista”


## Servizio Stateful

È stato implementato un **contatore condiviso**:

- stato mantenuto nel Registry  
- tutti i worker vi accedono  
- dimostra condivisione di stato tra repliche

---

## Come eseguire

### 1. Avvio Registry

```
go run registry.go
```

### 2. Avvio Worker

```
go run worker.go 6001 1
go run worker.go 6002 4
go run worker.go 6003 2
```

### 3. Client interattivo

```
go run client.go
```

Oppure demo automatica:

- doppio click su `start_demo.bat`

---

## Dimostrazione funzionalità

Dal menu del client:

1. Echo → mostra caching  
2. Add → mostra round robin  
3. Inc → mostra stato condiviso  
4. Reload → dimostra dinamicità

Nei log del registry è visibile:

- registrazione worker  
- peso assegnato  
- numero server attivi

---

## Scelte implementative

- Linguaggio: **Go**
- Middleware: **gRPC**

---

## Algoritmi di Load Balancing

### Stateless – Round Robin
Il client seleziona ciclicamente i server:

```
s = servers[index % len(servers)]
```

Non considera lo stato.

### Stateful – Weighted
La probabilità di scelta è proporzionale al peso:

- peso alto → più richieste  
- simula server con più risorse

---


## Architettura

Il sistema è composto da tre componenti:

1. **Registry**
   - mantiene l’elenco dei server attivi  
   - espone:
     - Register
     - GetServers
     - Inc (contatore condiviso)

2. **Worker**
   - espone due RPC semplici:
     - Echo(string) → string  
     - Add(a,b) → int  
   - inoltra al registry la chiamata `Inc`

3. **Client**
   - esegue lookup SOLO all’inizio sessione  
   - implementa:
     - Round Robin (stateless)
     - Weighted (stateful)
   - caching delle risposte

---
