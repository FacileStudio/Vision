import type { Row } from '../row';

/** Strings shared by the whole app: buttons, navigation, the sign-in page and the landing page. */
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
		'Lightweight, self-hosted web analytics for your projects.',
		'Des statistiques web légères et auto-hébergées pour vos projets.',
		'Analítica web ligera y autoalojada para tus proyectos.',
		'Schlanke, selbst gehostete Webanalyse für Ihre Projekte.'
	],
	'login.failedTitle': ['Sign-in failed', 'Connexion impossible', 'No se pudo iniciar sesión', 'Anmeldung fehlgeschlagen'],
	'login.failed': [
		'Facile could not sign you in. Try again.',
		'Facile n’a pas pu vous connecter. Réessayez.',
		'Facile no pudo iniciar tu sesión. Inténtalo de nuevo.',
		'Facile konnte Sie nicht anmelden. Versuchen Sie es erneut.'
	],

	'landing.heroTitle': [
		'See your traffic.|Know your audience.',
		'Voyez votre trafic.|Connaissez votre audience.',
		'Mira tu tráfico.|Conoce a tu audiencia.',
		'Sehen Sie Ihren Traffic.|Kennen Sie Ihr Publikum.'
	],
	'landing.heroText': [
		'Lightweight analytics for your sites. One script tag, live numbers.',
		'Des statistiques légères pour vos sites. Une balise script, des chiffres en direct.',
		'Analítica ligera para tus sitios. Una etiqueta script, cifras en directo.',
		'Schlanke Webanalyse für Ihre Websites. Ein Script-Tag, Zahlen in Echtzeit.'
	],
	'landing.demo': [
		'A Vision dashboard: visitors arrive, the counters roll and the current hour grows.',
		'Un tableau de bord Vision. Les visiteurs arrivent, les compteurs défilent et l’heure en cours grandit.',
		'Un panel de Vision. Llegan visitantes, los contadores suben y la hora actual crece.',
		'Ein Vision-Dashboard. Besucher kommen, die Zähler laufen und die aktuelle Stunde wächst.'
	],
	'landing.feature.script.title': ['One script tag', 'Une seule balise script', 'Una sola etiqueta script', 'Ein einziges Script-Tag'],
	'landing.feature.script.text': [
		'Add one tag to your site. No keys, nothing to set up.',
		'Ajoutez une balise à votre site. Pas de clé, rien à configurer.',
		'Añade una etiqueta a tu sitio. Sin claves, nada que configurar.',
		'Ein Tag auf Ihrer Website. Keine Schlüssel, nichts einzurichten.'
	],
	'landing.feature.realtime.title': ['Real-time analytics', 'Statistiques en temps réel', 'Analítica en tiempo real', 'Echtzeit-Analyse'],
	'landing.feature.realtime.text': [
		'Visitors, pages, sources, devices and countries on one dashboard.',
		'Visiteurs, pages, sources, appareils et pays sur un seul tableau de bord.',
		'Visitantes, páginas, fuentes, dispositivos y países en un solo panel.',
		'Besucher, Seiten, Quellen, Geräte und Länder in einem Dashboard.'
	],
	'landing.feature.goals.title': ['Goals', 'Objectifs', 'Objetivos', 'Ziele'],
	'landing.feature.goals.text': [
		'Count signups or purchases with one vision.track call.',
		'Comptez inscriptions ou achats avec un appel à vision.track.',
		'Cuenta registros o compras con una llamada a vision.track.',
		'Zählen Sie Anmeldungen oder Käufe mit einem Aufruf von vision.track.'
	],
	'landing.feature.locked.title': ['Domain-locked', 'Verrouillé par domaine', 'Bloqueado por dominio', 'An die Domain gebunden'],
	'landing.feature.locked.text': [
		'Vision checks each event against your domains, so nobody inflates your numbers.',
		'Vision vérifie chaque événement selon vos domaines. Personne ne gonfle vos chiffres.',
		'Vision comprueba cada evento con tus dominios, así nadie infla tus cifras.',
		'Vision prüft jedes Ereignis gegen Ihre Domains, damit niemand Ihre Zahlen aufbläht.'
	],
	'landing.pricing': ['Pricing', 'Tarifs', 'Precios', 'Preise'],
	'landing.tagline': [
		'Privacy-friendly web analytics.',
		'Des statistiques web respectueuses de la vie privée.',
		'Analítica web que respeta la privacidad.',
		'Datenschutzfreundliche Webanalyse.'
	],
	'landing.ctaTitle': ['Start counting today', 'Commencez à compter aujourd’hui', 'Empieza a contar hoy', 'Starten Sie noch heute'],
	'landing.ctaText': [
		'Sign in with your Facile account and add your first site.',
		'Connectez-vous avec votre compte Facile et ajoutez votre premier site.',
		'Inicia sesión con tu cuenta de Facile y añade tu primer sitio.',
		'Melden Sie sich mit Ihrem Facile-Konto an und fügen Sie Ihre erste Website hinzu.'
	]
} satisfies Record<string, Row>;
