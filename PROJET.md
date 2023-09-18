# Identification des exigences et des pré-requis

Pour la conception du systeme de paiement autonome

Exigences:
- API accessible de l'exterieur pour l'initiation du paiement
- API accessible aux services internes
- Gestion des paiements (initiation, validation, notification, statut)
- Gestion des api key des services ou application qui devrait utilisé le module pour plus de sécuriter
- Workflow de paiement (mobile money, carte bancaire, espèce, etc)
- Tableau de bord

Pré-requis:
- Disposé des api des maisons de téléphonies ou d'un agregateur de paiement mobile money
- Un serveur accessible de l'exterieur (pour les clients externe et interne)
- Un serveur de base de donnée (MySQL|MariaDB|PostgreSQL|SQLServer)

# Architecture

Pour ce projet nous adopterons une architecture simplifié avec:
- Serveur Web (Nginx)
- Backend (API Rest, Golang)
- Front-End Dashboard (ReactJs)
- Base de donnée (PostgreSQL)

Le serveur web fera office de service proxy et de load balancing pour la gestion des montées en charge. Le programme principale sera ecris en Golang. Il portera la logique metier. Nous nous appuyerons sur une base de donnée relationnel PostgreSQL pour sa capacité à gérer des gros trafic tout en etant rapide et open source.
Nous utiliserons ReactJs pour la partie graphique, ReactJs est un framework moderne qui s'adapte parfaitement aux usage moderne d'internet.

La communication se fera via des protocoles sécurisé telque: TLS, HTTPS.

# Developpement

# Formation et transfert de compétence

# Pipeline de déploiement

