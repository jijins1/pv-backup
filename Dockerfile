# Utilisez une image de base avec Go préinstallé
FROM golang:1.20 AS builder

# Définir le répertoire de travail à l'intérieur du conteneur
WORKDIR /app

# Copiez le code source de votre application Go dans le conteneur
COPY . .

# Construisez l'exécutable de votre application Go
RUN CGO_ENABLED=0 GOOS=linux go build



# Utilisez une image plus légère pour exécuter l'application, comme une image scratch
FROM scratch

# Copiez l'exécutable de votre application Go depuis le premier étape de construction
COPY --from=builder /app/pv-backup /pv-backup


# Définir le point d'entrée de votre application
ENTRYPOINT ["/pv-backup"]