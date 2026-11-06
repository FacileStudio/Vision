import type { Row } from '../row';

/** The site list, the add-site drawer and the ranked-list drawer. */
export const sites = {
	'sites.pageTitle': ['Sites — Vision', 'Sites — Vision', 'Sitios — Vision', 'Websites — Vision'],
	'sites.heading': ['Sites', 'Sites', 'Sitios', 'Websites'],
	'sites.subtitle': [
		'Everything {space} is tracking.',
		'Tout ce que suit {space}.',
		'Todo lo que sigue {space}.',
		'Alles, was {space} verfolgt.'
	],
	'sites.thisSpace': ['this space', 'cet espace', 'este espacio', 'dieser Bereich'],
	'sites.add': ['Add site', 'Ajouter un site', 'Añadir sitio', 'Website hinzufügen'],
	'sites.loadFailed': [
		'Could not load your sites.',
		'Impossible de charger vos sites.',
		'No se pudieron cargar tus sitios.',
		'Ihre Websites konnten nicht geladen werden.'
	],
	'sites.deleted': ['Deleted {name}.', '{name} a été supprimé.', 'Se eliminó {name}.', '{name} wurde gelöscht.'],
	'sites.tracked': [
		'{name} is now being tracked.',
		'{name} est maintenant suivi.',
		'Ahora se está siguiendo {name}.',
		'{name} wird jetzt erfasst.'
	],
	'sites.empty.title': ['No sites yet', 'Aucun site pour le moment', 'Aún no hay sitios', 'Noch keine Websites'],
	'sites.empty.text': [
		'Add a domain and drop one script tag on it. Numbers start arriving on the next pageview.',
		'Ajoutez un domaine et placez-y une balise script. Les chiffres arrivent dès la prochaine page vue.',
		'Añade un dominio y pon una etiqueta script en él. Las cifras empiezan a llegar con la siguiente visita.',
		'Fügen Sie eine Domain hinzu und binden Sie ein Script-Tag ein. Die Zahlen erscheinen ab dem nächsten Seitenaufruf.'
	],
	'sites.deleteAria': ['Delete {name}', 'Supprimer {name}', 'Eliminar {name}', '{name} löschen'],
	'sites.deleteTitle': ['Delete {name}?', 'Supprimer {name} ?', '¿Eliminar {name}?', '{name} löschen?'],
	'sites.thisSite': ['this site', 'ce site', 'este sitio', 'diese Website'],
	'sites.thisDomain': ['this domain', 'ce domaine', 'este dominio', 'diese Domain'],
	'sites.deleteBody': [
		'Every pageview, session and goal recorded for {domain} goes with it, and the script on the site stops being accepted. This cannot be undone.',
		'Toutes les pages vues, sessions et tous les objectifs enregistrés pour {domain} disparaissent avec lui, et le script du site n’est plus accepté. Cette action est irréversible.',
		'Todas las visitas, sesiones y objetivos registrados de {domain} se eliminan con él, y el script del sitio deja de aceptarse. No se puede deshacer.',
		'Alle für {domain} erfassten Seitenaufrufe, Sitzungen und Ziele werden mitgelöscht, und das Script auf der Website wird nicht mehr akzeptiert. Das lässt sich nicht rückgängig machen.'
	],
	'sites.deleteConfirm': ['Delete site', 'Supprimer le site', 'Eliminar sitio', 'Website löschen'],

	'addSite.title': ['Add a site', 'Ajouter un site', 'Añadir un sitio', 'Website hinzufügen'],
	'addSite.description': [
		'Vision starts counting as soon as the script is on the page.',
		'Vision commence à compter dès que le script est sur la page.',
		'Vision empieza a contar en cuanto el script está en la página.',
		'Vision beginnt zu zählen, sobald das Script auf der Seite eingebunden ist.'
	],
	'addSite.space': ['Space', 'Espace', 'Espacio', 'Bereich'],
	'addSite.spaceHelper': [
		'Everyone in the space can read this site’s stats.',
		'Tous les membres de l’espace peuvent lire les statistiques de ce site.',
		'Todos los miembros del espacio pueden leer las estadísticas de este sitio.',
		'Alle Mitglieder des Bereichs können die Statistiken dieser Website lesen.'
	],
	'addSite.nameHelper': [
		'How it appears in your site list.',
		'Son nom dans votre liste de sites.',
		'Cómo aparece en tu lista de sitios.',
		'So erscheint sie in Ihrer Website-Liste.'
	],
	'addSite.namePlaceholder': ['My website', 'Mon site web', 'Mi sitio web', 'Meine Website'],
	'addSite.domain': ['Domain', 'Domaine', 'Dominio', 'Domain'],
	'addSite.domainHelper': [
		'Events are verified against this domain — no subdomain, no scheme.',
		'Les événements sont vérifiés par rapport à ce domaine : sans sous-domaine ni protocole.',
		'Los eventos se verifican con este dominio: sin subdominio ni protocolo.',
		'Ereignisse werden gegen diese Domain geprüft: ohne Subdomain und ohne Protokoll.'
	],
	'addSite.noSpace': [
		'No space available — create one first.',
		'Aucun espace disponible : créez-en un d’abord.',
		'No hay ningún espacio disponible: crea uno primero.',
		'Kein Bereich verfügbar – legen Sie zuerst einen an.'
	],
	'addSite.failed': [
		'Could not add the site.',
		'Impossible d’ajouter le site.',
		'No se pudo añadir el sitio.',
		'Die Website konnte nicht hinzugefügt werden.'
	],
	'addSite.errorTitle': ['Not added', 'Non ajouté', 'No se añadió', 'Nicht hinzugefügt'],

	'drawer.summary': [
		{ one: '{total} total · {count} entry', other: '{total} total · {count} entries' },
		{ one: '{total} au total · {count} entrée', other: '{total} au total · {count} entrées' },
		{ one: '{total} en total · {count} entrada', other: '{total} en total · {count} entradas' },
		{ one: '{total} gesamt · {count} Eintrag', other: '{total} gesamt · {count} Einträge' }
	],
	'drawer.searchAria': ['Search {title}', 'Rechercher dans {title}', 'Buscar en {title}', '{title} durchsuchen'],
	'drawer.noMatch': [
		'Nothing matches “{query}”.',
		'Rien ne correspond à « {query} ».',
		'Nada coincide con «{query}».',
		'Nichts passt zu „{query}“.'
	]
} satisfies Record<string, Row>;
