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
		'Vision is a lightweight, self-hosted analytics tracker for your projects. One script tag, zero config, full visibility.',
		'Vision est un outil de statistiques léger et auto-hébergé pour vos projets. Une balise script, aucune configuration, une visibilité complète.',
		'Vision es una herramienta de analítica ligera y autoalojada para tus proyectos. Una etiqueta script, cero configuración, visibilidad total.',
		'Vision ist ein schlankes, selbst gehostetes Analyse-Tool für Ihre Projekte. Ein Script-Tag, keine Konfiguration, volle Transparenz.'
	],
	'landing.feature.script.title': ['One script tag', 'Une seule balise script', 'Una sola etiqueta script', 'Ein einziges Script-Tag'],
	'landing.feature.script.text': [
		'Drop a single <script> tag on your site. No API keys, no config, nothing to maintain.',
		'Ajoutez une seule balise <script> à votre site. Pas de clé d’API, pas de configuration, rien à maintenir.',
		'Añade una sola etiqueta <script> a tu sitio. Sin claves de API, sin configuración, nada que mantener.',
		'Fügen Sie Ihrer Website ein einziges <script>-Tag hinzu. Keine API-Schlüssel, keine Konfiguration, kein Wartungsaufwand.'
	],
	'landing.feature.realtime.title': ['Real-time analytics', 'Statistiques en temps réel', 'Analítica en tiempo real', 'Echtzeit-Analyse'],
	'landing.feature.realtime.text': [
		'Pageviews, visitors, entry and exit pages, referrers, devices and countries — one dashboard.',
		'Pages vues, visiteurs, pages d’entrée et de sortie, sources, appareils et pays : un seul tableau de bord.',
		'Visitas, visitantes, páginas de entrada y salida, fuentes, dispositivos y países: un solo panel.',
		'Seitenaufrufe, Besucher, Ein- und Ausstiegsseiten, Quellen, Geräte und Länder: alles in einem Dashboard.'
	],
	'landing.feature.locked.title': ['Domain-locked', 'Verrouillé par domaine', 'Bloqueado por dominio', 'An die Domain gebunden'],
	'landing.feature.locked.text': [
		'Events are verified against the domains you registered, so nobody can inflate your numbers.',
		'Les événements sont vérifiés par rapport aux domaines que vous avez déclarés : personne ne peut gonfler vos chiffres.',
		'Los eventos se verifican con los dominios que registraste, así que nadie puede inflar tus cifras.',
		'Ereignisse werden gegen die von Ihnen registrierten Domains geprüft, sodass niemand Ihre Zahlen verfälschen kann.'
	],
	'landing.pricing': ['Pricing', 'Tarifs', 'Precios', 'Preise'],
	'landing.tagline': [
		'Privacy-friendly, self-hosted web analytics.',
		'Des statistiques web respectueuses de la vie privée, auto-hébergées.',
		'Analítica web autoalojada que respeta la privacidad.',
		'Datenschutzfreundliche, selbst gehostete Webanalyse.'
	],
	'landing.ctaTitle': ['Ready to sign in?', 'Prêt à vous connecter ?', '¿Listo para iniciar sesión?', 'Bereit zur Anmeldung?'],
	'landing.ctaText': [
		'Use your Facile account to reach Vision.',
		'Utilisez votre compte Facile pour accéder à Vision.',
		'Usa tu cuenta de Facile para acceder a Vision.',
		'Nutzen Sie Ihr Facile-Konto, um Vision zu öffnen.'
	]
} satisfies Record<string, Row>;
