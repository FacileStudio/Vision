import type { Row } from '../row';

/** The page where the person confirms what an AI client asked Vision to do. */
export const pending = {
	'pending.title': ['Pending actions — Vision', 'Actions en attente — Vision', 'Acciones pendientes — Vision', 'Ausstehende Aktionen — Vision'],
	'pending.heading': ['Pending actions', 'Actions en attente', 'Acciones pendientes', 'Ausstehende Aktionen'],
	'pending.intro': [
		'An AI client asked Vision to do these. Nothing happens until you confirm.',
		'Un client IA a demandé ces actions à Vision. Rien ne se passe tant que vous ne confirmez pas.',
		'Un cliente de IA pidió estas acciones a Vision. No ocurre nada hasta que confirmes.',
		'Ein KI-Client hat Vision um diese Aktionen gebeten. Nichts passiert, bevor Sie bestätigen.'
	],
	'pending.none': ['Nothing is waiting', 'Rien en attente', 'Nada pendiente', 'Nichts ausstehend'],
	'pending.noneBody': [
		'Actions an AI client asks for appear here until you confirm or dismiss them.',
		'Les actions demandées par un client IA apparaissent ici jusqu’à ce que vous les confirmiez ou les ignoriez.',
		'Las acciones que pide un cliente de IA aparecen aquí hasta que las confirmes o descartes.',
		'Aktionen, um die ein KI-Client bittet, erscheinen hier, bis Sie sie bestätigen oder verwerfen.'
	],
	'pending.kind.delete_site': ['Delete site', 'Supprimer le site', 'Eliminar sitio', 'Website löschen'],
	'pending.kind.share_site': ['Public link', 'Lien public', 'Enlace público', 'Öffentlicher Link'],
	'pending.kind.delete_goal': ['Delete goal', 'Supprimer l’objectif', 'Eliminar objetivo', 'Ziel löschen'],
	'pending.askedBy': ['Asked by {client}', 'Demandé par {client}', 'Pedido por {client}', 'Angefragt von {client}'],
	'pending.note.delete_site': [
		'The site and its dashboard are deleted for good.',
		'Le site et son tableau de bord sont supprimés définitivement.',
		'El sitio y su panel se eliminan para siempre.',
		'Die Website und ihr Dashboard werden endgültig gelöscht.'
	],
	'pending.note.share_site': [
		'Anyone with the link can see this site’s dashboard without signing in. An older link stops working.',
		'Toute personne qui a le lien voit le tableau de bord de ce site sans se connecter. Un ancien lien cesse de fonctionner.',
		'Cualquiera con el enlace ve el panel de este sitio sin iniciar sesión. Un enlace anterior deja de funcionar.',
		'Jeder mit dem Link sieht das Dashboard dieser Website ohne Anmeldung. Ein älterer Link funktioniert nicht mehr.'
	],
	'pending.note.delete_goal': [
		'The goal is deleted for good. Its visits stay in your analytics.',
		'L’objectif est supprimé définitivement. Ses visites restent dans vos statistiques.',
		'El objetivo se elimina para siempre. Sus visitas siguen en tus estadísticas.',
		'Das Ziel wird endgültig gelöscht. Die Besuche bleiben in Ihren Statistiken.'
	],
	'pending.confirm.delete_site': ['Delete the site', 'Supprimer le site', 'Eliminar el sitio', 'Website löschen'],
	'pending.confirm.share_site': ['Open the link', 'Ouvrir le lien', 'Abrir el enlace', 'Link öffnen'],
	'pending.confirm.delete_goal': ['Delete the goal', 'Supprimer l’objectif', 'Eliminar el objetivo', 'Ziel löschen'],
	'pending.goalOn': ['Goal on {site}', 'Objectif sur {site}', 'Objetivo en {site}', 'Ziel auf {site}'],
	'pending.dismiss': ['Dismiss', 'Ignorer', 'Descartar', 'Verwerfen'],
	'pending.changed': [
		'This changed after the request. Ask the AI client again.',
		'Cet élément a changé après la demande. Redemandez au client IA.',
		'Esto cambió después de la solicitud. Vuelve a pedírselo al cliente de IA.',
		'Dies wurde nach der Anfrage geändert. Fragen Sie den KI-Client erneut.'
	],
	'pending.missing': [
		'This no longer exists, or you can no longer change it.',
		'Cet élément n’existe plus, ou vous ne pouvez plus le modifier.',
		'Esto ya no existe, o ya no puedes cambiarlo.',
		'Dies existiert nicht mehr, oder Sie können es nicht mehr ändern.'
	],
	'pending.done.deleted': ['Deleted', 'Supprimé', 'Eliminado', 'Gelöscht'],
	'pending.done.shared': ['Public link open', 'Lien public ouvert', 'Enlace público abierto', 'Öffentlicher Link geöffnet'],
	'pending.copy': ['Copy link', 'Copier le lien', 'Copiar enlace', 'Link kopieren'],
	'pending.copied': ['Link copied', 'Lien copié', 'Enlace copiado', 'Link kopiert'],
	'pending.close': ['Done', 'Terminé', 'Listo', 'Fertig'],
	'pending.dismissed': ['Request dismissed', 'Demande ignorée', 'Solicitud descartada', 'Anfrage verworfen'],
	'pending.loadFailed': [
		'Could not load the pending actions.',
		'Impossible de charger les actions en attente.',
		'No se pudieron cargar las acciones pendientes.',
		'Die ausstehenden Aktionen konnten nicht geladen werden.'
	],
	'pending.signInAgain.title': [
		'Sign in again to confirm',
		'Reconnectez-vous pour confirmer',
		'Vuelve a iniciar sesión para confirmar',
		'Melden Sie sich erneut an, um zu bestätigen'
	],
	'pending.signInAgain.body': [
		'Only a fresh sign-in in your browser can confirm what an AI client asked. Sign in again here, then come back to this page. The mobile and desktop apps cannot confirm.',
		'Seule une connexion récente dans votre navigateur peut confirmer ce qu’un client IA a demandé. Reconnectez-vous ici, puis revenez sur cette page. Les applications mobile et de bureau ne peuvent pas confirmer.',
		'Solo un inicio de sesión reciente en tu navegador puede confirmar lo que pidió un cliente de IA. Vuelve a iniciar sesión aquí y regresa a esta página. Las aplicaciones móvil y de escritorio no pueden confirmar.',
		'Nur eine frische Anmeldung in Ihrem Browser kann bestätigen, worum ein KI-Client gebeten hat. Melden Sie sich hier erneut an und kehren Sie dann zu dieser Seite zurück. Die Mobil- und Desktop-Apps können nicht bestätigen.'
	],
	'pending.signInAgain.action': ['Sign in again', 'Se reconnecter', 'Iniciar sesión de nuevo', 'Erneut anmelden'],
	'pending.failed': [
		'That did not work. Try again.',
		'Cela n’a pas fonctionné. Réessayez.',
		'No funcionó. Inténtalo de nuevo.',
		'Das hat nicht geklappt. Versuchen Sie es erneut.'
	]
} as const satisfies Record<string, Row>;
