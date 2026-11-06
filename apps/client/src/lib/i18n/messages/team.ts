import type { Row } from '../row';

/** Spaces: the list, the creation page and the space settings. */
export const team = {
	'role.owner': ['Owner', 'Propriétaire', 'Propietario', 'Eigentümer'],
	'role.admin': ['Admin', 'Administrateur', 'Administrador', 'Administrator'],
	'role.editor': ['Editor', 'Éditeur', 'Editor', 'Redakteur'],
	'role.viewer': ['Viewer', 'Lecteur', 'Lector', 'Betrachter'],

	'team.pageTitle': ['Teams — Vision', 'Équipes — Vision', 'Equipos — Vision', 'Teams — Vision'],
	'team.heading': ['Teams', 'Équipes', 'Equipos', 'Teams'],
	'team.subtitle': [
		'Spaces you belong to. Everyone in a space sees its sites and their stats.',
		'Les espaces dont vous faites partie. Tous les membres d’un espace voient ses sites et leurs statistiques.',
		'Los espacios a los que perteneces. Todos los miembros de un espacio ven sus sitios y sus estadísticas.',
		'Bereiche, denen Sie angehören. Alle Mitglieder eines Bereichs sehen dessen Websites und Statistiken.'
	],
	'team.manageRegistre': ['Manage in Registre', 'Gérer dans Registre', 'Gestionar en Registre', 'In Registre verwalten'],
	'team.newSpace': ['New space', 'Nouvel espace', 'Nuevo espacio', 'Neuer Bereich'],
	'team.loadFailed': [
		'Could not load your spaces.',
		'Impossible de charger vos espaces.',
		'No se pudieron cargar tus espacios.',
		'Ihre Bereiche konnten nicht geladen werden.'
	],
	'team.searchPlaceholder': ['Search spaces…', 'Rechercher un espace…', 'Buscar espacios…', 'Bereiche suchen…'],
	'team.searchAria': ['Search spaces', 'Rechercher un espace', 'Buscar espacios', 'Bereiche suchen'],
	'team.noMatch': [
		'No space matches “{query}”',
		'Aucun espace ne correspond à « {query} »',
		'Ningún espacio coincide con «{query}»',
		'Kein Bereich passt zu „{query}“'
	],
	'team.noMatchHint': [
		'Try another name, or clear the search.',
		'Essayez un autre nom ou effacez la recherche.',
		'Prueba con otro nombre o borra la búsqueda.',
		'Versuchen Sie einen anderen Namen oder löschen Sie die Suche.'
	],
	'team.empty': ['No spaces yet', 'Aucun espace pour le moment', 'Aún no hay espacios', 'Noch keine Bereiche'],
	'team.emptyHint': [
		'A space holds sites and the people who can read them.',
		'Un espace regroupe des sites et les personnes qui peuvent les consulter.',
		'Un espacio reúne sitios y las personas que pueden leerlos.',
		'Ein Bereich bündelt Websites und die Personen, die sie lesen können.'
	],
	'team.members': [
		{ one: '{count} member', other: '{count} members' },
		{ one: '{count} membre', other: '{count} membres' },
		{ one: '{count} miembro', other: '{count} miembros' },
		{ one: '{count} Mitglied', other: '{count} Mitglieder' }
	],
	'team.sites': [
		{ one: '{count} site', other: '{count} sites' },
		{ one: '{count} site', other: '{count} sites' },
		{ one: '{count} sitio', other: '{count} sitios' },
		{ one: '{count} Website', other: '{count} Websites' }
	],

	'newSpace.pageTitle': ['New space — Vision', 'Nouvel espace — Vision', 'Nuevo espacio — Vision', 'Neuer Bereich — Vision'],
	'newSpace.back': ['Back to teams', 'Retour aux équipes', 'Volver a equipos', 'Zurück zu den Teams'],
	'newSpace.heading': ['New space', 'Nouvel espace', 'Nuevo espacio', 'Neuer Bereich'],
	'newSpace.text': [
		'A space groups sites and the people who can read them. You will be its owner.',
		'Un espace regroupe des sites et les personnes qui peuvent les consulter. Vous en serez le propriétaire.',
		'Un espacio agrupa sitios y las personas que pueden leerlos. Serás su propietario.',
		'Ein Bereich bündelt Websites und die Personen, die sie lesen können. Sie sind der Eigentümer.'
	],
	'newSpace.nameHelper': [
		'Usually a client, a team or a product.',
		'En général un client, une équipe ou un produit.',
		'Normalmente un cliente, un equipo o un producto.',
		'Meist ein Kunde, ein Team oder ein Produkt.'
	],
	'newSpace.namePlaceholder': ['Acme Studio', 'Studio Acme', 'Estudio Acme', 'Acme Studio'],
	'newSpace.tooShort': ['At least two characters.', 'Au moins deux caractères.', 'Al menos dos caracteres.', 'Mindestens zwei Zeichen.'],
	'newSpace.tooShortLong': [
		'A space name needs at least two characters.',
		'Le nom d’un espace doit comporter au moins deux caractères.',
		'El nombre de un espacio necesita al menos dos caracteres.',
		'Der Name eines Bereichs braucht mindestens zwei Zeichen.'
	],
	'newSpace.ready': ['{name} is ready.', '{name} est prêt.', '{name} está listo.', '{name} ist bereit.'],
	'newSpace.failed': [
		'Could not create the space.',
		'Impossible de créer l’espace.',
		'No se pudo crear el espacio.',
		'Der Bereich konnte nicht erstellt werden.'
	],
	'newSpace.errorTitle': ['Not created', 'Non créé', 'No se creó', 'Nicht erstellt'],
	'newSpace.submit': ['Create space', 'Créer l’espace', 'Crear espacio', 'Bereich erstellen'],

	'space.fallbackTitle': ['Space', 'Espace', 'Espacio', 'Bereich'],
	'space.loadFailed': [
		'Could not load this space.',
		'Impossible de charger cet espace.',
		'No se pudo cargar este espacio.',
		'Dieser Bereich konnte nicht geladen werden.'
	],
	'space.openFailed': [
		'Could not open this space',
		'Impossible d’ouvrir cet espace',
		'No se pudo abrir este espacio',
		'Dieser Bereich konnte nicht geöffnet werden'
	],
	'space.nameAria': ['Space name', 'Nom de l’espace', 'Nombre del espacio', 'Name des Bereichs'],
	'space.renameAria': ['Rename space', 'Renommer l’espace', 'Cambiar el nombre del espacio', 'Bereich umbenennen'],
	'space.renamed': ['Space renamed.', 'Espace renommé.', 'Espacio renombrado.', 'Bereich umbenannt.'],
	'space.renameFailed': [
		'Could not rename the space.',
		'Impossible de renommer l’espace.',
		'No se pudo renombrar el espacio.',
		'Der Bereich konnte nicht umbenannt werden.'
	],
	'space.membersTitle': ['Members', 'Membres', 'Miembros', 'Mitglieder'],
	'space.membersManaged': [
		'Members and their roles are managed in Registre, for every Facile app at once.',
		'Les membres et leurs rôles sont gérés dans Registre, pour toutes les apps Facile à la fois.',
		'Los miembros y sus roles se gestionan en Registre, para todas las apps de Facile a la vez.',
		'Mitglieder und ihre Rollen werden in Registre verwaltet, für alle Facile-Apps gleichzeitig.'
	],
	'space.membersLocal': [
		'Owners cannot be removed, and nobody can change their own role.',
		'Les propriétaires ne peuvent pas être retirés, et personne ne peut modifier son propre rôle.',
		'Los propietarios no se pueden quitar y nadie puede cambiar su propio rol.',
		'Eigentümer können nicht entfernt werden, und niemand kann seine eigene Rolle ändern.'
	],
	'space.manageLabel': ['Manage this space', 'Gérer cet espace', 'Gestionar este espacio', 'Diesen Bereich verwalten'],
	'space.manageText': [
		'Invite people, change roles or leave in Registre.',
		'Invitez des personnes, changez les rôles ou quittez l’espace dans Registre.',
		'Invita a personas, cambia roles o abandona el espacio en Registre.',
		'Laden Sie Personen ein, ändern Sie Rollen oder verlassen Sie den Bereich in Registre.'
	],
	'space.roleAria': ['Role for {name}', 'Rôle de {name}', 'Rol de {name}', 'Rolle von {name}'],
	'space.removeAria': ['Remove {name}', 'Retirer {name}', 'Quitar a {name}', '{name} entfernen'],
	'space.removeTitle': ['Remove {name}?', 'Retirer {name} ?', '¿Quitar a {name}?', '{name} entfernen?'],
	'space.thisMember': ['this member', 'ce membre', 'este miembro', 'dieses Mitglied'],
	'space.removeBody': [
		'They lose access to every site in this space immediately. The stats they looked at are unaffected — this only revokes access.',
		'Cette personne perd immédiatement l’accès à tous les sites de cet espace. Les statistiques qu’elle consultait ne changent pas : seul l’accès est révoqué.',
		'Pierde el acceso a todos los sitios de este espacio de inmediato. Las estadísticas que consultaba no cambian: solo se revoca el acceso.',
		'Die Person verliert sofort den Zugriff auf alle Websites dieses Bereichs. Die Statistiken selbst bleiben unverändert – es wird nur der Zugriff entzogen.'
	],
	'space.removeKeep': ['Keep access', 'Conserver l’accès', 'Mantener el acceso', 'Zugriff behalten'],
	'space.removed': [
		'{name} no longer has access.',
		'{name} n’a plus accès.',
		'{name} ya no tiene acceso.',
		'{name} hat keinen Zugriff mehr.'
	],
	'space.addTitle': ['Add someone', 'Ajouter quelqu’un', 'Añadir a alguien', 'Person hinzufügen'],
	'space.addText': [
		'They need a Vision account already — adding grants access immediately.',
		'Cette personne doit déjà avoir un compte Vision : l’ajout donne accès immédiatement.',
		'Necesita tener ya una cuenta de Vision: al añadirla se concede el acceso de inmediato.',
		'Die Person braucht bereits ein Vision-Konto – das Hinzufügen gewährt sofort Zugriff.'
	],
	'space.addEmailPlaceholder': ['name@studio.com', 'nom@studio.com', 'nombre@estudio.com', 'name@studio.com'],
	'space.added': ['{email} was added.', '{email} a été ajouté.', 'Se añadió a {email}.', '{email} wurde hinzugefügt.'],
	'space.addFailed': [
		'Could not add that person.',
		'Impossible d’ajouter cette personne.',
		'No se pudo añadir a esa persona.',
		'Die Person konnte nicht hinzugefügt werden.'
	],
	'space.addErrorTitle': ['Not added', 'Non ajouté', 'No se añadió', 'Nicht hinzugefügt'],
	'space.roleUpdated': ['Role updated.', 'Rôle mis à jour.', 'Rol actualizado.', 'Rolle aktualisiert.'],
	'space.roleFailed': [
		'Could not change that role.',
		'Impossible de modifier ce rôle.',
		'No se pudo cambiar ese rol.',
		'Die Rolle konnte nicht geändert werden.'
	],
	'space.dangerTitle': ['Danger zone', 'Zone dangereuse', 'Zona de peligro', 'Gefahrenbereich'],
	'space.dangerText': [
		'Irreversible, and nobody can undo it for you.',
		'Irréversible, et personne ne peut l’annuler à votre place.',
		'Irreversible, y nadie puede deshacerlo por ti.',
		'Nicht rückgängig zu machen, und niemand kann es für Sie zurücknehmen.'
	],
	'space.deleteLabel': ['Delete this space', 'Supprimer cet espace', 'Eliminar este espacio', 'Diesen Bereich löschen'],
	'space.deleteText': [
		'Every site must be removed first. Members lose access the moment it goes.',
		'Tous les sites doivent d’abord être supprimés. Les membres perdent l’accès dès qu’il disparaît.',
		'Primero hay que eliminar todos los sitios. Los miembros pierden el acceso en cuanto se elimina.',
		'Zuerst müssen alle Websites entfernt werden. Die Mitglieder verlieren den Zugriff, sobald er gelöscht ist.'
	],
	'space.deleteButton': ['Delete space', 'Supprimer l’espace', 'Eliminar espacio', 'Bereich löschen'],
	'space.deleteTitle': ['Delete {name}?', 'Supprimer {name} ?', '¿Eliminar {name}?', '{name} löschen?'],
	'space.deleteBody': [
		'The space and its membership are gone for good. Sites must already be removed.',
		'L’espace et ses membres disparaissent définitivement. Les sites doivent déjà être supprimés.',
		'El espacio y sus miembros desaparecen para siempre. Los sitios ya deben haberse eliminado.',
		'Der Bereich und seine Mitglieder sind endgültig weg. Die Websites müssen bereits entfernt sein.'
	],
	'space.deleteType': [
		'Type DELETE to confirm',
		'Saisissez DELETE pour confirmer',
		'Escribe DELETE para confirmar',
		'Geben Sie zur Bestätigung DELETE ein'
	],
	'space.deleted': ['Space deleted.', 'Espace supprimé.', 'Espacio eliminado.', 'Bereich gelöscht.'],
	'space.leaveLabel': ['Leave this space', 'Quitter cet espace', 'Abandonar este espacio', 'Diesen Bereich verlassen'],
	'space.leaveText': [
		'You lose access to every site in it. An owner can add you back.',
		'Vous perdez l’accès à tous ses sites. Un propriétaire peut vous réintégrer.',
		'Pierdes el acceso a todos sus sitios. Un propietario puede volver a añadirte.',
		'Sie verlieren den Zugriff auf alle ihre Websites. Ein Eigentümer kann Sie wieder hinzufügen.'
	],
	'space.leaveButton': ['Leave space', 'Quitter l’espace', 'Abandonar espacio', 'Bereich verlassen'],
	'space.leaveTitle': ['Leave {name}?', 'Quitter {name} ?', '¿Abandonar {name}?', '{name} verlassen?'],
	'space.thisSpace': ['this space', 'cet espace', 'este espacio', 'diesen Bereich'],
	'space.leaveBody': [
		'You lose access to every site in it, and only an owner can add you back.',
		'Vous perdez l’accès à tous ses sites, et seul un propriétaire peut vous réintégrer.',
		'Pierdes el acceso a todos sus sitios y solo un propietario puede volver a añadirte.',
		'Sie verlieren den Zugriff auf alle ihre Websites, und nur ein Eigentümer kann Sie wieder hinzufügen.'
	],
	'space.leaveStay': ['Stay', 'Rester', 'Quedarme', 'Bleiben'],
	'space.left': ['You left the space.', 'Vous avez quitté l’espace.', 'Has abandonado el espacio.', 'Sie haben den Bereich verlassen.']
} satisfies Record<string, Row>;
