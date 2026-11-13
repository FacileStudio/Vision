import type { Row } from '../row';

/** Settings: profile, appearance, API keys and report webhooks. */
export const settings = {
	'settings.pageTitle': ['Settings — Vision', 'Réglages — Vision', 'Ajustes — Vision', 'Einstellungen — Vision'],
	'settings.heading': ['Settings', 'Réglages', 'Ajustes', 'Einstellungen'],
	'settings.subtitle': [
		'Your account, this browser, and everything wired to your spaces.',
		'Votre compte, ce navigateur et tout ce qui est relié à vos espaces.',
		'Tu cuenta, este navegador y todo lo conectado a tus espacios.',
		'Ihr Konto, dieser Browser und alles, was mit Ihren Bereichen verbunden ist.'
	],
	'settings.sectionsLabel': ['Settings sections', 'Sections des réglages', 'Secciones de ajustes', 'Einstellungsbereiche'],
	'settings.profile': ['Profile', 'Profil', 'Perfil', 'Profil'],
	'settings.preferences': ['Preferences', 'Préférences', 'Preferencias', 'Einstellungen'],
	'settings.notifications': ['Notifications', 'Notifications', 'Notificaciones', 'Benachrichtigungen'],
	'settings.tokens': ['API tokens', 'Jetons d’API', 'Tokens de API', 'API-Tokens'],



	'apiKeys.endpointTitle': ['Endpoint', 'Point d’accès', 'Endpoint', 'Endpunkt'],
	'apiKeys.endpointText': [
		'Point a script or an HTTP client here. Not a secret — the key is.',
		'Dirigez un script ou un client HTTP ici. Ce n’est pas un secret : la clé, si.',
		'Apunta un script o un cliente HTTP aquí. No es un secreto: la clave sí.',
		'Richten Sie ein Script oder einen HTTP-Client hierauf. Das ist kein Geheimnis – der Schlüssel schon.'
	],
	'apiKeys.baseUrl': ['Base URL', 'URL de base', 'URL base', 'Basis-URL'],
	'apiKeys.loadFailed': [
		'Could not load your API keys.',
		'Impossible de charger vos clés d’API.',
		'No se pudieron cargar tus claves de API.',
		'Ihre API-Schlüssel konnten nicht geladen werden.'
	],
	'apiKeys.scopeRead': ['read', 'lecture', 'lectura', 'Lesen'],
	'apiKeys.scopeWrite': ['read + write', 'lecture + écriture', 'lectura + escritura', 'Lesen + Schreiben'],
	'apiKeys.revoke': ['Revoke', 'Révoquer', 'Revocar', 'Widerrufen'],
	'apiKeys.revoked': ['Key revoked.', 'Clé révoquée.', 'Clave revocada.', 'Schlüssel widerrufen.'],
	'apiKeys.scope': ['Scope', 'Portée', 'Alcance', 'Berechtigung'],
	'apiKeys.scopeReadOption': [
		'Read — fetch analytics',
		'Lecture : consulter les statistiques',
		'Lectura: consultar las estadísticas',
		'Lesen – Statistiken abrufen'
	],
	'apiKeys.scopeWriteOption': [
		'Read and write — fetch and change data',
		'Lecture et écriture : consulter et modifier les données',
		'Lectura y escritura: consultar y modificar datos',
		'Lesen und Schreiben – Daten abrufen und ändern'
	],
	'settings.space.title': ['Space', 'Espace', 'Espacio', 'Bereich'],
	'settings.space.text': [
		'API keys and webhooks belong to one space. Pick the one to manage.',
		'Les clés d’API et les webhooks appartiennent à un espace. Choisissez celui à gérer.',
		'Las claves de API y los webhooks pertenecen a un espacio. Elige cuál gestionar.',
		'API-Schlüssel und Webhooks gehören zu einem Bereich. Wählen Sie den zu verwaltenden.'
	]
} satisfies Record<string, Row>;
