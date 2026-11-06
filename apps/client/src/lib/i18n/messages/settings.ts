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
	'settings.appearance': ['Appearance', 'Apparence', 'Apariencia', 'Darstellung'],
	'settings.notifications': ['Notifications', 'Notifications', 'Notificaciones', 'Benachrichtigungen'],
	'settings.api': ['API', 'API', 'API', 'API'],

	'profile.loadFailed': [
		'Could not load your profile.',
		'Impossible de charger votre profil.',
		'No se pudo cargar tu perfil.',
		'Ihr Profil konnte nicht geladen werden.'
	],
	'profile.memberSince': ['Member since', 'Membre depuis', 'Miembro desde', 'Mitglied seit'],
	'profile.avatar': ['Avatar', 'Avatar', 'Avatar', 'Avatar'],
	'profile.avatarSso': [
		'Managed in single sign-on',
		'Géré par l’authentification unique',
		'Gestionado por el inicio de sesión único',
		'Über die einmalige Anmeldung verwaltet'
	],
	'profile.avatarInitials': ['Initials', 'Initiales', 'Iniciales', 'Initialen'],
	'profile.accountTitle': ['Account', 'Compte', 'Cuenta', 'Konto'],
	'profile.accountText': [
		'How you appear to everyone else in your spaces.',
		'Comment vous apparaissez aux autres membres de vos espaces.',
		'Cómo te ven los demás miembros de tus espacios.',
		'Wie Sie für alle anderen in Ihren Bereichen erscheinen.'
	],
	'profile.displayName': ['Display name', 'Nom affiché', 'Nombre visible', 'Anzeigename'],
	'profile.displayNameHelper': [
		'Shown to the other members of your spaces.',
		'Affiché aux autres membres de vos espaces.',
		'Se muestra a los demás miembros de tus espacios.',
		'Wird den anderen Mitgliedern Ihrer Bereiche angezeigt.'
	],
	'profile.namePlaceholder': ['Your name', 'Votre nom', 'Tu nombre', 'Ihr Name'],
	'profile.emailManaged': [
		'Managed by your Facile account — change it there.',
		'Géré par votre compte Facile : modifiez-le là-bas.',
		'Gestionado por tu cuenta de Facile: cámbialo allí.',
		'Wird über Ihr Facile-Konto verwaltet – ändern Sie sie dort.'
	],
	'profile.emailHelper': [
		'Used to be added to a space.',
		'Sert à vous ajouter à un espace.',
		'Se usa para que te añadan a un espacio.',
		'Dient dazu, zu einem Bereich hinzugefügt zu werden.'
	],
	'profile.save': ['Save changes', 'Enregistrer les modifications', 'Guardar cambios', 'Änderungen speichern'],
	'profile.updated': ['Profile updated.', 'Profil mis à jour.', 'Perfil actualizado.', 'Profil aktualisiert.'],
	'profile.saveFailed': [
		'Could not save your profile.',
		'Impossible d’enregistrer votre profil.',
		'No se pudo guardar tu perfil.',
		'Ihr Profil konnte nicht gespeichert werden.'
	],
	'profile.sessionTitle': ['Session', 'Session', 'Sesión', 'Sitzung'],
	'profile.sessionText': ['This device only.', 'Cet appareil uniquement.', 'Solo este dispositivo.', 'Nur dieses Gerät.'],
	'profile.logout': ['Log out', 'Se déconnecter', 'Cerrar sesión', 'Abmelden'],
	'profile.logoutText': [
		'Ends this session here. Your other devices stay signed in.',
		'Termine la session ici. Vos autres appareils restent connectés.',
		'Termina la sesión aquí. Tus otros dispositivos siguen conectados.',
		'Beendet die Sitzung hier. Ihre anderen Geräte bleiben angemeldet.'
	],

	'appearance.themeTitle': ['Theme', 'Thème', 'Tema', 'Design'],
	'appearance.themeText': [
		'Applied to this browser. Every device you sign in from keeps its own choice.',
		'Appliqué à ce navigateur. Chaque appareil depuis lequel vous vous connectez garde son propre choix.',
		'Se aplica a este navegador. Cada dispositivo desde el que inicias sesión conserva su propia elección.',
		'Gilt für diesen Browser. Jedes Gerät, mit dem Sie sich anmelden, behält seine eigene Wahl.'
	],
	'appearance.scheme': ['Colour scheme', 'Thème de couleurs', 'Esquema de color', 'Farbschema'],
	'appearance.schemeText': [
		'System follows your OS for as long as the app is open.',
		'« Système » suit votre système d’exploitation tant que l’app est ouverte.',
		'«Sistema» sigue tu sistema operativo mientras la app esté abierta.',
		'„System“ folgt Ihrem Betriebssystem, solange die App geöffnet ist.'
	],
	'appearance.system': ['System', 'Système', 'Sistema', 'System'],
	'appearance.light': ['Light', 'Clair', 'Claro', 'Hell'],
	'appearance.dark': ['Dark', 'Sombre', 'Oscuro', 'Dunkel'],

	'apiKeys.endpointTitle': ['Endpoint', 'Point d’accès', 'Endpoint', 'Endpunkt'],
	'apiKeys.endpointText': [
		'Point a script or an HTTP client here. Not a secret — the key is.',
		'Dirigez un script ou un client HTTP ici. Ce n’est pas un secret : la clé, si.',
		'Apunta un script o un cliente HTTP aquí. No es un secreto: la clave sí.',
		'Richten Sie ein Script oder einen HTTP-Client hierauf. Das ist kein Geheimnis – der Schlüssel schon.'
	],
	'apiKeys.baseUrl': ['Base URL', 'URL de base', 'URL base', 'Basis-URL'],
	'apiKeys.title': ['API keys', 'Clés d’API', 'Claves de API', 'API-Schlüssel'],
	'apiKeys.text': [
		'One key per machine. Revoking one never touches the others.',
		'Une clé par machine. En révoquer une n’affecte jamais les autres.',
		'Una clave por máquina. Revocar una nunca afecta a las demás.',
		'Ein Schlüssel pro Rechner. Das Widerrufen eines Schlüssels berührt die anderen nie.'
	],
	'apiKeys.new': ['New key', 'Nouvelle clé', 'Nueva clave', 'Neuer Schlüssel'],
	'apiKeys.loadFailed': [
		'Could not load your API keys.',
		'Impossible de charger vos clés d’API.',
		'No se pudieron cargar tus claves de API.',
		'Ihre API-Schlüssel konnten nicht geladen werden.'
	],
	'apiKeys.empty': [
		'No keys yet. A script needs one before it can read this space’s analytics.',
		'Aucune clé pour le moment. Un script en a besoin pour lire les statistiques de cet espace.',
		'Aún no hay claves. Un script necesita una para leer las estadísticas de este espacio.',
		'Noch keine Schlüssel. Ein Script braucht einen, um die Statistiken dieses Bereichs zu lesen.'
	],
	'apiKeys.colName': ['Name', 'Nom', 'Nombre', 'Name'],
	'apiKeys.colKey': ['Key', 'Clé', 'Clave', 'Schlüssel'],
	'apiKeys.colScope': ['Scope', 'Portée', 'Alcance', 'Berechtigung'],
	'apiKeys.colLastUsed': ['Last used', 'Dernier usage', 'Último uso', 'Zuletzt verwendet'],
	'apiKeys.colActions': ['Actions', 'Actions', 'Acciones', 'Aktionen'],
	'apiKeys.scopeRead': ['read', 'lecture', 'lectura', 'Lesen'],
	'apiKeys.scopeWrite': ['read + write', 'lecture + écriture', 'lectura + escritura', 'Lesen + Schreiben'],
	'apiKeys.revokedBadge': ['Revoked', 'Révoquée', 'Revocada', 'Widerrufen'],
	'apiKeys.revoke': ['Revoke', 'Révoquer', 'Revocar', 'Widerrufen'],
	'apiKeys.revokeAria': ['Revoke {name}', 'Révoquer {name}', 'Revocar {name}', '{name} widerrufen'],
	'apiKeys.revokeTitle': [
		'Revoke this key?',
		'Révoquer cette clé ?',
		'¿Revocar esta clave?',
		'Diesen Schlüssel widerrufen?'
	],
	'apiKeys.revokeBody': [
		'“{name}” stops working immediately, and anything still using it starts failing. It cannot be un-revoked — the row stays listed so the audit trail still names it.',
		'« {name} » cesse de fonctionner immédiatement, et tout ce qui l’utilise encore échoue. La révocation est définitive : la ligne reste listée pour que l’historique la nomme toujours.',
		'«{name}» deja de funcionar de inmediato y todo lo que aún lo use empezará a fallar. No se puede deshacer: la fila sigue en la lista para que el registro la nombre.',
		'„{name}“ funktioniert sofort nicht mehr, und alles, was ihn noch nutzt, schlägt fehl. Das Widerrufen lässt sich nicht rückgängig machen – die Zeile bleibt aufgelistet, damit das Protokoll sie weiterhin nennt.'
	],
	'apiKeys.revokeConfirm': ['Revoke key', 'Révoquer la clé', 'Revocar clave', 'Schlüssel widerrufen'],
	'apiKeys.revoked': ['Key revoked.', 'Clé révoquée.', 'Clave revocada.', 'Schlüssel widerrufen.'],
	'apiKeys.drawerTitle': ['New API key', 'Nouvelle clé d’API', 'Nueva clave de API', 'Neuer API-Schlüssel'],
	'apiKeys.copyTitle': ['Copy it now', 'Copiez-la maintenant', 'Cópiala ahora', 'Jetzt kopieren'],
	'apiKeys.copyText': [
		'This is the only time the key is shown. Vision stores a hash, so it cannot be shown to you again — losing it means issuing a new one.',
		'C’est la seule fois que la clé est affichée. Vision n’en conserve qu’un hachage et ne peut donc plus vous la montrer : la perdre oblige à en émettre une nouvelle.',
		'Esta es la única vez que se muestra la clave. Vision guarda un hash, así que no puede volver a mostrártela: perderla obliga a emitir una nueva.',
		'Dies ist das einzige Mal, dass der Schlüssel angezeigt wird. Vision speichert nur einen Hash und kann ihn Ihnen nicht erneut zeigen – wenn Sie ihn verlieren, müssen Sie einen neuen ausstellen.'
	],
	'apiKeys.secretHelper': [
		'Store it in your password manager or your CI secret store.',
		'Conservez-la dans votre gestionnaire de mots de passe ou le coffre de secrets de votre CI.',
		'Guárdala en tu gestor de contraseñas o en el almacén de secretos de tu CI.',
		'Bewahren Sie ihn in Ihrem Passwortmanager oder im Secret-Speicher Ihrer CI auf.'
	],
	'apiKeys.nameHelper': [
		'Where the key will live — a machine, a pipeline, a script.',
		'Où vivra la clé : une machine, un pipeline, un script.',
		'Dónde vivirá la clave: una máquina, un pipeline, un script.',
		'Wo der Schlüssel verwendet wird – ein Rechner, eine Pipeline, ein Script.'
	],
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
	'apiKeys.createFailed': [
		'Could not create the key.',
		'Impossible de créer la clé.',
		'No se pudo crear la clave.',
		'Der Schlüssel konnte nicht erstellt werden.'
	],
	'apiKeys.notCreated': ['Not created', 'Non créée', 'No se creó', 'Nicht erstellt'],
	'apiKeys.submit': ['Create key', 'Créer la clé', 'Crear clave', 'Schlüssel erstellen']
} satisfies Record<string, Row>;
