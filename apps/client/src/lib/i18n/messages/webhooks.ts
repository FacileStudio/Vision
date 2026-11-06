import type { Row } from '../row';

/** Report webhooks. */
export const webhooks = {
	'webhooks.title': ['Report webhooks', 'Webhooks de rapports', 'Webhooks de informes', 'Bericht-Webhooks'],
	'webhooks.text': [
		'Vision posts a signed analytics summary to each URL on its own schedule. Nothing else is emailed or sent.',
		'Vision envoie un résumé signé des statistiques à chaque URL selon son propre calendrier. Rien d’autre n’est envoyé par e-mail ou autrement.',
		'Vision envía un resumen firmado de las estadísticas a cada URL según su propio calendario. No se envía nada más por correo ni de otro modo.',
		'Vision sendet an jede URL nach eigenem Zeitplan eine signierte Zusammenfassung der Statistiken. Sonst wird nichts per E-Mail oder anderweitig verschickt.'
	],
	'webhooks.new': ['New webhook', 'Nouveau webhook', 'Nuevo webhook', 'Neuer Webhook'],
	'webhooks.loadFailed': [
		'Could not load your report webhooks.',
		'Impossible de charger vos webhooks de rapports.',
		'No se pudieron cargar tus webhooks de informes.',
		'Ihre Bericht-Webhooks konnten nicht geladen werden.'
	],
	'webhooks.empty': [
		'No webhooks yet. Add one to receive a periodic summary of every site in this space.',
		'Aucun webhook pour le moment. Ajoutez-en un pour recevoir un résumé périodique de chaque site de cet espace.',
		'Aún no hay webhooks. Añade uno para recibir un resumen periódico de cada sitio de este espacio.',
		'Noch keine Webhooks. Fügen Sie einen hinzu, um regelmäßig eine Zusammenfassung jeder Website dieses Bereichs zu erhalten.'
	],
	'webhooks.hourly': ['Hourly', 'Toutes les heures', 'Cada hora', 'Stündlich'],
	'webhooks.every6': ['Every 6 hours', 'Toutes les 6 heures', 'Cada 6 horas', 'Alle 6 Stunden'],
	'webhooks.every12': ['Every 12 hours', 'Toutes les 12 heures', 'Cada 12 horas', 'Alle 12 Stunden'],
	'webhooks.daily': ['Daily', 'Tous les jours', 'Cada día', 'Täglich'],
	'webhooks.weekly': ['Weekly', 'Toutes les semaines', 'Cada semana', 'Wöchentlich'],
	'webhooks.monthly': ['Monthly', 'Tous les mois', 'Cada mes', 'Monatlich'],
	'webhooks.everyHours': [
		{ one: 'Every hour', other: 'Every {count} hours' },
		{ one: 'Toutes les heures', other: 'Toutes les {count} heures' },
		{ one: 'Cada hora', other: 'Cada {count} horas' },
		{ one: 'Jede Stunde', other: 'Alle {count} Stunden' }
	],
	'webhooks.everyDays': [
		{ one: 'Every day', other: 'Every {count} days' },
		{ one: 'Tous les jours', other: 'Tous les {count} jours' },
		{ one: 'Cada día', other: 'Cada {count} días' },
		{ one: 'Jeden Tag', other: 'Alle {count} Tage' }
	],
	'webhooks.custom': ['Custom…', 'Personnalisée…', 'Personalizada…', 'Benutzerdefiniert…'],
	'webhooks.lastSent': [
		'{frequency} · last sent {date}',
		'{frequency} · dernier envoi le {date}',
		'{frequency} · último envío el {date}',
		'{frequency} · zuletzt gesendet am {date}'
	],
	'webhooks.neverSent': [
		'{frequency} · never sent',
		'{frequency} · jamais envoyé',
		'{frequency} · nunca enviado',
		'{frequency} · noch nie gesendet'
	],
	'webhooks.active': ['Active', 'Actif', 'Activo', 'Aktiv'],
	'webhooks.paused': ['Paused', 'En pause', 'En pausa', 'Pausiert'],
	'webhooks.enableAria': ['Enable {url}', 'Activer {url}', 'Activar {url}', '{url} aktivieren'],
	'webhooks.testAria': [
		'Send a test report to {url}',
		'Envoyer un rapport de test à {url}',
		'Enviar un informe de prueba a {url}',
		'Testbericht an {url} senden'
	],
	'webhooks.editAria': ['Edit {url}', 'Modifier {url}', 'Editar {url}', '{url} bearbeiten'],
	'webhooks.deleteAria': ['Delete {url}', 'Supprimer {url}', 'Eliminar {url}', '{url} löschen'],
	'webhooks.toggleFailed': [
		'Could not change that webhook.',
		'Impossible de modifier ce webhook.',
		'No se pudo cambiar ese webhook.',
		'Der Webhook konnte nicht geändert werden.'
	],
	'webhooks.testSent': ['Test report sent.', 'Rapport de test envoyé.', 'Informe de prueba enviado.', 'Testbericht gesendet.'],
	'webhooks.testFailed': [
		'The test delivery failed.',
		'L’envoi de test a échoué.',
		'El envío de prueba falló.',
		'Die Testzustellung ist fehlgeschlagen.'
	],
	'webhooks.updated': ['Webhook updated.', 'Webhook mis à jour.', 'Webhook actualizado.', 'Webhook aktualisiert.'],
	'webhooks.added': ['Webhook added.', 'Webhook ajouté.', 'Webhook añadido.', 'Webhook hinzugefügt.'],
	'webhooks.saveFailed': [
		'Could not save the webhook.',
		'Impossible d’enregistrer le webhook.',
		'No se pudo guardar el webhook.',
		'Der Webhook konnte nicht gespeichert werden.'
	],
	'webhooks.notSaved': ['Not saved', 'Non enregistré', 'No se guardó', 'Nicht gespeichert'],
	'webhooks.deleted': ['Webhook deleted.', 'Webhook supprimé.', 'Webhook eliminado.', 'Webhook gelöscht.'],
	'webhooks.editTitle': ['Edit webhook', 'Modifier le webhook', 'Editar webhook', 'Webhook bearbeiten'],
	'webhooks.drawerText': [
		'Reports are signed with the shared secret so the receiver can verify them.',
		'Les rapports sont signés avec le secret partagé pour que le destinataire puisse les vérifier.',
		'Los informes se firman con el secreto compartido para que el receptor pueda verificarlos.',
		'Berichte werden mit dem gemeinsamen Geheimnis signiert, damit der Empfänger sie prüfen kann.'
	],
	'webhooks.url': ['URL', 'URL', 'URL', 'URL'],
	'webhooks.urlHelper': [
		'Where the report is POSTed.',
		'L’adresse à laquelle le rapport est envoyé en POST.',
		'Dónde se envía el informe con POST.',
		'Wohin der Bericht per POST gesendet wird.'
	],
	'webhooks.secret': ['Shared secret', 'Secret partagé', 'Secreto compartido', 'Gemeinsames Geheimnis'],
	'webhooks.secretKeep': [
		'Leave empty to keep the secret you already set.',
		'Laissez vide pour conserver le secret déjà défini.',
		'Déjalo vacío para conservar el secreto ya definido.',
		'Leer lassen, um das bereits gesetzte Geheimnis zu behalten.'
	],
	'webhooks.secretNew': [
		'Used to sign every delivery. Store it wherever the receiver reads it from.',
		'Sert à signer chaque envoi. Conservez-le là où le destinataire le lit.',
		'Se usa para firmar cada envío. Guárdalo donde lo lea el receptor.',
		'Dient zum Signieren jeder Zustellung. Bewahren Sie es dort auf, wo der Empfänger es liest.'
	],
	'webhooks.frequency': ['Frequency', 'Fréquence', 'Frecuencia', 'Häufigkeit'],
	'webhooks.everyInHours': ['Every, in hours', 'Toutes les, en heures', 'Cada, en horas', 'Alle, in Stunden'],
	'webhooks.update': ['Update', 'Mettre à jour', 'Actualizar', 'Aktualisieren'],
	'webhooks.add': ['Add webhook', 'Ajouter le webhook', 'Añadir webhook', 'Webhook hinzufügen'],
	'webhooks.deleteTitle': [
		'Delete this webhook?',
		'Supprimer ce webhook ?',
		'¿Eliminar este webhook?',
		'Diesen Webhook löschen?'
	],
	'webhooks.deleteIt': ['it', 'cette adresse', 'esa dirección', 'diese Adresse'],
	'webhooks.deleteBody': [
		'Nothing is posted to {url} again, and the schedule is not recoverable. Pause it instead if you only want the reports to stop for a while.',
		'Plus rien n’est envoyé à {url}, et le calendrier ne peut pas être restauré. Mettez-le plutôt en pause si vous voulez seulement arrêter les rapports un temps.',
		'No se vuelve a enviar nada a {url} y el calendario no se puede recuperar. Mejor ponlo en pausa si solo quieres detener los informes un tiempo.',
		'An {url} wird nichts mehr gesendet, und der Zeitplan lässt sich nicht wiederherstellen. Pausieren Sie ihn stattdessen, wenn die Berichte nur eine Weile ausbleiben sollen.'
	],
	'webhooks.deleteConfirm': ['Delete webhook', 'Supprimer le webhook', 'Eliminar webhook', 'Webhook löschen']
} satisfies Record<string, Row>;
