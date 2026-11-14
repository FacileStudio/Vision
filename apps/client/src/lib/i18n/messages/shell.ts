import type { Row } from '../row';

/** Strings shared by the whole app: buttons, navigation and the sign-in page. */
export const shell = {
	'common.cancel': ['Cancel', 'Annuler', 'Cancelar', 'Abbrechen'],
	'common.save': ['Save', 'Enregistrer', 'Guardar', 'Speichern'],
	'common.saving': ['Saving…', 'Enregistrement…', 'Guardando…', 'Wird gespeichert…'],
	'common.add': ['Add', 'Ajouter', 'Añadir', 'Hinzufügen'],
	'common.adding': ['Adding…', 'Ajout…', 'Añadiendo…', 'Wird hinzugefügt…'],
	'common.creating': ['Creating…', 'Création…', 'Creando…', 'Wird erstellt…'],
	'common.delete': ['Delete', 'Supprimer', 'Eliminar', 'Löschen'],
	'common.remove': ['Remove', 'Retirer', 'Quitar', 'Entfernen'],
	'common.done': ['Done', 'Terminé', 'Listo', 'Fertig'],
	'common.keepIt': ['Keep it', 'Conserver', 'Conservar', 'Behalten'],
	'common.loading': ['Loading…', 'Chargement…', 'Cargando…', 'Wird geladen…'],
	'common.search': ['Search…', 'Rechercher…', 'Buscar…', 'Suchen…'],
	'common.seeAll': ['See all', 'Tout voir', 'Ver todo', 'Alle anzeigen'],
	'common.clear': ['Clear', 'Effacer', 'Borrar', 'Zurücksetzen'],
	'common.never': ['Never', 'Jamais', 'Nunca', 'Nie'],
	'common.you': ['you', 'vous', 'tú', 'Sie'],
	'common.name': ['Name', 'Nom', 'Nombre', 'Name'],
	'common.email': ['Email', 'E-mail', 'Correo electrónico', 'E-Mail'],
	'common.role': ['Role', 'Rôle', 'Rol', 'Rolle'],
	'common.nothingYet': ['Nothing recorded yet.', 'Rien d’enregistré pour le moment.', 'Todavía no hay nada registrado.', 'Noch nichts erfasst.'],

	'nav.sites': ['Sites', 'Sites', 'Sitios', 'Websites'],
	'nav.teams': ['Teams', 'Équipes', 'Equipos', 'Teams'],
	'nav.account': ['Account', 'Compte', 'Cuenta', 'Konto'],

	'login.title': ['Log in — Vision', 'Connexion — Vision', 'Iniciar sesión — Vision', 'Anmelden — Vision'],
	'login.tagline': [
		'Your site’s visitors, without cookies.',
		'Les visites de votre site, sans cookies.',
		'Las visitas de tu sitio, sin cookies.',
		'Die Besuche Ihrer Website, ohne Cookies.'
	],
	'login.failedTitle': ['Sign-in failed', 'Connexion impossible', 'No se pudo iniciar sesión', 'Anmeldung fehlgeschlagen'],
	'login.failed': [
		'Facile could not sign you in. Try again.',
		'Facile n’a pas pu vous connecter. Réessayez.',
		'Facile no pudo iniciar tu sesión. Inténtalo de nuevo.',
		'Facile konnte Sie nicht anmelden. Versuchen Sie es erneut.'
	],
} satisfies Record<string, Row>;
