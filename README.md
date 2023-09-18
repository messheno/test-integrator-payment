# test-integrator-payment
Projet de test

# Identification des exigences et des pré-requis

Pour ce projet de mise en place de solution de paiement autonome pour Sikem Assurance, voici les exigences et prerequis.

En ce qui concerne les exigences nous en avons 2.
A savoir: les exigences fonctionnelles et les exigences non fonctionnelles.

## Les exigences fonctionnelles

1. Création de paiement (API):
    - L'API doit permettre aux applications externes d'initier des paiements en fournissant les informations suivantes : montant, description, références de paiement et références de l'application qui initie le paiement.
    - L'API doit générer un identifiant unique pour chaque transaction de paiement.
2. Validation des données (API) :
    - L'API doit valider le montant à payer pour s'assurer qu'il est valide et positif.
    - L'API doit vérifier que le numéro de téléphone fourni est au format valide.
    - L'API doit s'assurer que le mode de paiement est pris en charge par la solution.
3. Gestion des paiements en cours (API) :
    - L'API doit maintenir une liste des paiements en cours avec des informations telles que l'identifiant de transaction, le montant, la date et l'heure, le statut et les références de l'application.
    - L'API doit permettre de mettre à jour le statut des paiements en cours (en attente, réussi, échoué, annulé, etc.).
4. Communication avec les services internes (API) :
    - L'API doit être en mesure de transmettre toutes les transactions, y compris le montant, la date et l'heure, le mode de paiement et le statut, aux services internes de Sikem Assurance.
    - L'API doit fournir des tokens d'accès sécurisés pour permettre aux services internes d'accéder aux informations de paiement.

## Les exigences non fonctionnelles
1. Sécurité (API et Interface utilisateur):
    - Toutes les communications entre les applications externes, l'API et les services internes doivent être sécurisées via des protocoles de cryptage tels que HTTPS.
    - L'API doit avoir des mécanismes d'authentification robustes pour s'assurer que seules les applications autorisées peuvent l'utiliser.
    - Les données sensibles telles que les informations de paiement doivent être stockées de manière sécurisée, conformément aux normes de sécurité applicables.
2. Performance (API et Interface utilisateur):
    - L'API doit être capable de gérer un volume élevé de transactions en un temps raisonnable, avec un temps de réponse minimal.
    - L'interface utilisateur doit offrir une expérience utilisateur fluide même en cas de forte utilisation.
3. Disponibilité (API et Interface utilisateur):
    - L'API doit être hautement disponible pour garantir que les paiements peuvent être initiés à tout moment.
    - L'interface utilisateur doit également être disponible en permanence pour les agents de Sikem Assurance.
4. Scalabilité (API):
    - L'API doit être conçue pour être évolutive, capable de gérer une augmentation future du nombre de transactions.
5. Conformité aux réglementations (API et Interface utilisateur):
    - La solution doit être conforme aux réglementations locales et internationales relatives aux paiements en ligne, à la protection des données et à la sécurité des transactions.
6. Tests (API et Interface utilisateur):
    - Des tests d'acceptation, de sécurité et de performance doivent être effectués pour garantir le bon fonctionnement de la solution.

## Les prerequis
1. Infrastructure Technique:
    - Disponibilité d'une infrastructure de serveurs et de bases de données pour héberger l'API de paiement et l'interface utilisateur.
    - Capacité de mise en place de mécanismes de sauvegarde et de reprise en cas de panne pour garantir la disponibilité continue du système.
2. Ressources Humaines:
    - Avoir une équipe de développement logiciel qualifiée pour la création de l'API et de l'interface utilisateur.
    - Disposer d'une équipe de test pour effectuer des tests de validation et de sécurité.
    - Avoir des experts en sécurité de l'information pour garantir la sécurité des transactions.
3. Budget:
    - Disposer d'un budget alloué pour le développement, les tests, la maintenance et la documentation de la solution.
    - Prévoir des ressources pour les coûts d'exploitation continus, tels que l'hébergement, la sécurité et les mises à jour.
4. Exigences Légales et Réglementaires:
    - Identifier et comprendre les réglementations locales et internationales liées aux paiements en ligne et à la protection des données, et s'assurer que la solution est conforme à ces réglementations.
5. Accès aux Services Externes:
 - Si la solution doit communiquer avec des services externes, tels que des processeurs de paiement ou des banques, s'assurer d'avoir accès à ces services et de disposer des autorisations nécessaires.
6. Intégration aux Systèmes Internes:
    - Si la solution doit s'intégrer aux systèmes internes de Sikem Assurance, garantir que les interfaces et les API nécessaires sont disponibles et prêtes à être intégrées.
7. Plan de Gestion des Projets :
    - Élaborer un plan de gestion de projet détaillé qui inclut les étapes de développement, de tests, de déploiement et de maintenance, ainsi que les ressources allouées à chaque étape.
8. Plan de Sécurité:
    - Élaborer un plan de sécurité qui décrit les mesures de sécurité à mettre en place pour protéger les données sensibles des transactions et garantir la confidentialité et l'intégrité des paiements.
9. Communications avec les Parties Prenantes:
    - Établir des canaux de communication avec toutes les parties prenantes, y compris les équipes internes, les régulateurs, les fournisseurs de services externes et les utilisateurs finaux, pour assurer une coordination efficace.
10. Infrastructure de Surveillance:
    - Mettre en place une infrastructure de surveillance pour surveiller les performances, la disponibilité et la sécurité de la solution en temps réel.
11. Plan de Formation:
    - Élaborer un plan de formation pour les agents de Sikem Assurance afin de s'assurer qu'ils sont compétents pour utiliser l'interface utilisateur et gérer les transactions.

# Architecture

Cette architecture est basée sur les exigences fournies precedement et suppose que nous utilisons une architecture web moderne. Pour cette demo nous allons utilisé ces technologie:
- Golang (API)
- ReactJs (Frontend)
- PostgreSQL (Base de donnée)
- Keycloak envisager mais pas utilisé pour cette demo (Système de gestion d'identité et d'accès)

## Architecture technique

1. Frontend avec ReactJs
    - Nous utilisons ReactJs pour créer l'interface utilisateur permettant aux agents de Sikem Assurance de suivre les transactions.
    - Les agents se connectent à l'interface utilisateur via un navigateur web pour accéder aux fonctionnalités de gestion des transactions.
2. Backend avec Go
    - Nous utilisons Go pour créer l'API de paiement autonome qui permet aux applications externes d'initier des paiements et gère les paiements en cours.
    - Go est un excellent choix en raison de sa performance, de sa simplicité et de sa concurrence, ce qui le rend adapté à la gestion des transactions en temps réel.
3. Base de Données PostgreSQL
    - Nous utilisons PostgreSQL comme base de données pour stocker les informations sur les transactions en cours, les utilisateurs et les comptes.
    - PostgreSQL offre des fonctionnalités avancées de gestion des données, de sécurité et de performances.
4. Serveur Web avec Nginx
    - Nous utilisons Nginx pour servir l'application React.js et faire office de proxy inverse pour l'API Go.
    - Nginx peut également gérer la sécurité, la mise en cache et la gestion du trafic.
5. Infrastructure de Sécurité
    - Nous mettons en place des mécanismes de sécurité tels que HTTPS pour sécuriser les communications entre les agents, les applications externes, l'API et la base de données.
    - Nous implémentons l'authentification et l'autorisation pour garantir que seuls les utilisateurs autorisés peuvent accéder aux données et aux fonctionnalités de paiement.

## Architecture Fonctionnelle

1. Création de Paiement (API Go)
    - Les applications externes envoient des demandes HTTP à l'API Go pour créer des paiements en fournissant les détails nécessaires (montant, description, références, etc.).
    - L'API Go valide les données, génère un identifiant unique pour la transaction, puis stocke les informations dans la base de données PostgreSQL.
2. Gestion des Paiements en Cours (API Go)
    - L'API Go maintient une liste des paiements en cours dans la base de données, permettant aux applications externes de suivre le statut des paiements.
    - Les agents de Sikem Assurance utilisent l'interface utilisateur React.js pour accéder à ces informations et gérer les paiements.
3. Interface Utilisateur (React.js)
    - Les agents se connectent à l'interface utilisateur via un navigateur web pour visualiser, filtrer et gérer les transactions en cours.
    - L'interface utilisateur communique avec l'API Go via des appels HTTP pour obtenir les données de paiement.
4. Base de Données PostgreSQL
    - La base de données stocke les transactions, les informations sur les utilisateurs et les comptes, ce qui permet à l'API Go et à l'interface utilisateur de récupérer les données nécessaires.
5. Sécurité et Authentification
    - Toutes les communications sont sécurisées via HTTPS.
    - Les agents s'authentifient auprès de l'application React.js pour accéder aux fonctionnalités de gestion des paiements.
    - L'API Go effectue des vérifications d'authentification et d'autorisation pour les applications externes.

# Developpement
Les code sources:
- API (./api)
- FrontEnd (./front-end)

Chaque dossier contient un fichier dockerfile et docker-compose pour la contenerisation.

# Formation et transfert de compétence

# Pipeline de déploiement
