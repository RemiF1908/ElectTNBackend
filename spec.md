C'est un backend Go pour une application web de élections. Le but c'est de faire une api RESTful pour gérer les élections.
Il faut qu'une personne puisse s'authentifié avec de l'oauth2 (pour l'instant c'est google mais le projet c'est d'avoir son propre système d'oauth pour un sso)
Il faut qu'une personne authentifié puisse se présenter à une élection
Il faut qu'une personne authentifié puisse voter pour un candidat qui se présente à une élection
Il faut qu'un administrateur puisse créer une élection
Il faut qu'un administrateur puisse supprimer une élection
Il faut qu'un administrateur puisse modifier une élection
Il faut qu'un administrateur puisse voir les résultats d'une élection
Il faut qu'un administrateur puisse choisir si une election est ouverte ou fermée
Il faut qu'un administrateur puisse choisir si une election à du résulat en direct
Les élections sont toujours anonymes
Il ne faut garder aucun log des votes

Specification technique
Authentification : Gorilla/sessions, OAuth2 avec Google
Database : PostgreSQL
ORM : GORM
Framework : Echo
On va utiliser openapi avec oapi-codegen avec l'option strict server (@generate.sh)
Il faut donc commencer par crée le l'openapi et une fois l'openapi fait, on généra le code et on pourra commencer l'implémentations.
IL NE FAUT SURTOUT PAS ENREGISTRER QUI A VOTÉ POUR QUI, il faut juste garder le nombre de vote pour chaque candidat et qui a voté
Il faut essayer de séparer la business logic du code métier dans des services distincts
