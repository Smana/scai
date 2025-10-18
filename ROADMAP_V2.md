# Roadmap to v2

Cette première implémentation n'est pas idéal car elle prend en entrée des paramètre statiques et n'évolue pas.
De plus, l'utilisation d'une commande opentofu de façon impérative ne permet pas de s'assurer de la dérive dans le temps et d'observer le statut des ressources.

Tout d'abord je pense qu'il serait préférable d'utiliser l'API Kubernetes comme interface principale. Cela permettrait de bénéficier des fonctionnalités de base comme la réconciliation automatique. Je propose d'utiliser crossplane pour concevoir les ressources cloud. En effet, grâce aux compositions et surtout aux fonctions, nous pouvons aller bien plus loin, et implémenter les bonnes pratiques ou politiques décrites par l'utilisateur en custom ressources: voir exemple ici https://github.com/Smana/cloud-native-ref/tree/main/infrastructure/base/crossplane/configuration/kcl/app
Cela nécessite que le backend de notre application soit hébergé dans un cluster central qui servira de control plane. Puis les cloud providers des utilisateurs seront définis avec des ressources Kubernetes.

Pour le premier problème, nous pourrions implémenter un RAG permettant d'alimenter une base de données vectorielle et ainsi améliorer la pertinence des choix d'implémentation. Nous pourrons alimenter avec toutes les ressources kubernetes depuis l'api afin de donner un contexte précis au model.

Et scia devrait avant tout être une API dans kube pour effectuer les tâches faites actuellement en ligne de commande, (choisir le meilleure type d'api rest / graphql pour ce besoin) et un frontend rudimentaire comme celui de gemini par exemple pour y mettre son prompt.
