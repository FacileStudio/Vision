import type { Row } from '../row';

/** The site dashboard, the public dashboard and the charts and lists they share. */
export const dashboard = {
	'range.today': ['Today', 'Aujourd’hui', 'Hoy', 'Heute'],
	'range.7d': ['7d', '7 j', '7 d', '7 T'],
	'range.30d': ['30d', '30 j', '30 d', '30 T'],
	'range.90d': ['90d', '90 j', '90 d', '90 T'],
	'range.this_year': ['Year', 'Année', 'Año', 'Jahr'],
	'range.custom': ['Custom', 'Perso.', 'Personalizado', 'Eigener'],
	'range.label': ['Date range', 'Période', 'Intervalo de fechas', 'Zeitraum'],
	'range.from': ['From', 'Du', 'Desde', 'Von'],
	'range.to': ['To', 'Au', 'Hasta', 'Bis'],

	'granularity.label': ['Granularity', 'Granularité', 'Granularidad', 'Granularität'],
	'granularity.hour': ['Hourly', 'Par heure', 'Por hora', 'Stündlich'],
	'granularity.day': ['Daily', 'Par jour', 'Por día', 'Täglich'],
	'granularity.week': ['Weekly', 'Par semaine', 'Por semana', 'Wöchentlich'],
	'granularity.month': ['Monthly', 'Par mois', 'Por mes', 'Monatlich'],

	'site.loadFailed': [
		'Could not load this site.',
		'Impossible de charger ce site.',
		'No se pudo cargar este sitio.',
		'Diese Website konnte nicht geladen werden.'
	],
	'site.live': ['Live', 'En direct', 'En vivo', 'Live'],
	'site.reconnecting': ['Reconnecting…', 'Reconnexion…', 'Reconectando…', 'Verbindung wird wiederhergestellt…'],
	'site.export': ['Export', 'Exporter', 'Exportar', 'Exportieren'],
	'site.exportFailed': [
		'The export failed. Nothing was downloaded.',
		'L’export a échoué. Rien n’a été téléchargé.',
		'La exportación falló. No se descargó nada.',
		'Der Export ist fehlgeschlagen. Es wurde nichts heruntergeladen.'
	],

	'tracking.title': ['Tracking', 'Suivi', 'Seguimiento', 'Tracking'],
	'tracking.text': [
		'Paste this in the site’s <head>. Nothing else to configure.',
		'Collez ceci dans le <head> du site. Rien d’autre à configurer.',
		'Pega esto en el <head> del sitio. No hay nada más que configurar.',
		'Fügen Sie dies in den <head> der Website ein. Mehr ist nicht zu konfigurieren.'
	],
	'tracking.script': ['Script tag', 'Balise script', 'Etiqueta script', 'Script-Tag'],

	'share.title': ['Public dashboard', 'Tableau de bord public', 'Panel público', 'Öffentliches Dashboard'],
	'share.descriptionOn': [
		'Anyone with the link can read these numbers — and only these.',
		'Toute personne disposant du lien peut lire ces chiffres, et uniquement ceux-ci.',
		'Cualquiera con el enlace puede leer estas cifras, y solo estas.',
		'Jeder mit dem Link kann diese Zahlen lesen – und nur diese.'
	],
	'share.descriptionOff': [
		'Create a read-only link you can hand to a client without an account.',
		'Créez un lien en lecture seule à remettre à un client sans compte.',
		'Crea un enlace de solo lectura que puedes entregar a un cliente sin cuenta.',
		'Erstellen Sie einen Nur-Lese-Link, den Sie einem Kunden ohne Konto geben können.'
	],
	'share.link': ['Share link', 'Lien de partage', 'Enlace para compartir', 'Freigabelink'],
	'share.revoke': ['Revoke link', 'Révoquer le lien', 'Revocar enlace', 'Link widerrufen'],
	'share.create': ['Create public link', 'Créer un lien public', 'Crear enlace público', 'Öffentlichen Link erstellen'],
	'share.created': ['Public link created.', 'Lien public créé.', 'Enlace público creado.', 'Öffentlicher Link erstellt.'],
	'share.revoked': ['Public link revoked.', 'Lien public révoqué.', 'Enlace público revocado.', 'Öffentlicher Link widerrufen.'],
	'share.revokeTitle': [
		'Revoke the public link?',
		'Révoquer le lien public ?',
		'¿Revocar el enlace público?',
		'Öffentlichen Link widerrufen?'
	],
	'share.revokeBody': [
		'Anyone holding it gets a “not found” page from then on, and a new link will be a different URL. Nothing about the site’s data changes.',
		'Toute personne qui le détient verra une page « introuvable », et un nouveau lien aura une autre URL. Les données du site ne changent pas.',
		'Quien lo tenga verá una página de «no encontrado» a partir de ahora, y un enlace nuevo tendrá otra URL. Los datos del sitio no cambian.',
		'Wer ihn besitzt, sieht ab sofort eine Seite „Nicht gefunden“, und ein neuer Link hat eine andere URL. An den Daten der Website ändert sich nichts.'
	],

	'filters.buttonCount': ['Filters · {count}', 'Filtres · {count}', 'Filtros · {count}', 'Filter · {count}'],
	'filters.clear': ['Clear filters', 'Effacer les filtres', 'Borrar filtros', 'Filter zurücksetzen'],
	'filters.title': ['Filters', 'Filtres', 'Filtros', 'Filter'],
	'filters.description': [
		'Narrow every number on this page.',
		'Affinez tous les chiffres de cette page.',
		'Acota todas las cifras de esta página.',
		'Grenzen Sie alle Zahlen dieser Seite ein.'
	],
	'filters.country': ['Country', 'Pays', 'País', 'Land'],
	'filters.browser': ['Browser', 'Navigateur', 'Navegador', 'Browser'],
	'filters.os': ['OS', 'Système', 'SO', 'Betriebssystem'],
	'filters.device': ['Device', 'Appareil', 'Dispositivo', 'Gerät'],
	'filters.path': ['Path', 'Chemin', 'Ruta', 'Pfad'],
	'filters.referrer': ['Referrer', 'Source', 'Origen', 'Referrer'],

	'stat.pageviews': ['Pageviews', 'Pages vues', 'Visitas a páginas', 'Seitenaufrufe'],
	'stat.visitors': ['Visitors', 'Visiteurs', 'Visitantes', 'Besucher'],
	'stat.viewsPerVisitor': ['Views / visitor', 'Vues / visiteur', 'Vistas / visitante', 'Aufrufe / Besucher'],
	'stat.bounceRate': ['Bounce rate', 'Taux de rebond', 'Tasa de rebote', 'Absprungrate'],
	'stat.activeNow': ['Active now', 'Actifs maintenant', 'Activos ahora', 'Jetzt aktiv'],
	'stat.rightNow': ['right now', 'en ce moment', 'ahora mismo', 'gerade jetzt'],

	'traffic.title': ['Traffic over time', 'Trafic dans le temps', 'Tráfico a lo largo del tiempo', 'Traffic im Zeitverlauf'],
	'traffic.text': [
		'Pageviews and unique visitors.',
		'Pages vues et visiteurs uniques.',
		'Visitas a páginas y visitantes únicos.',
		'Seitenaufrufe und eindeutige Besucher.'
	],
	'traffic.sources': ['Traffic sources', 'Sources de trafic', 'Fuentes de tráfico', 'Traffic-Quellen'],
	'traffic.centerLabel': ['pageviews', 'pages vues', 'visitas', 'Seitenaufrufe'],
	'traffic.byHour': ['By hour of day', 'Par heure de la journée', 'Por hora del día', 'Nach Tageszeit'],
	'source.direct': ['Direct', 'Direct', 'Directo', 'Direkt'],
	'source.search': ['Search', 'Recherche', 'Búsqueda', 'Suche'],
	'source.social': ['Social', 'Réseaux sociaux', 'Redes sociales', 'Soziale Netzwerke'],
	'source.other': ['Other', 'Autres', 'Otros', 'Sonstige'],

	'goals.title': ['Goals', 'Objectifs', 'Objetivos', 'Ziele'],
	'goals.measured': [
		{ one: 'Measured against {count} unique visitor.', other: 'Measured against {count} unique visitors.' },
		{ one: 'Mesuré sur {count} visiteur unique.', other: 'Mesuré sur {count} visiteurs uniques.' },
		{ one: 'Medido sobre {count} visitante único.', other: 'Medido sobre {count} visitantes únicos.' },
		{ one: 'Gemessen an {count} eindeutigem Besucher.', other: 'Gemessen an {count} eindeutigen Besuchern.' }
	],
	'goals.hint': [
		'Conversion rates for the pages and events that matter.',
		'Taux de conversion des pages et des événements qui comptent.',
		'Tasas de conversión de las páginas y los eventos que importan.',
		'Conversionsraten für die Seiten und Ereignisse, auf die es ankommt.'
	],
	'goals.add': ['Add goal', 'Ajouter un objectif', 'Añadir objetivo', 'Ziel hinzufügen'],
	'goals.row': [
		{ one: '{count} conversion · {rate} of visitors', other: '{count} conversions · {rate} of visitors' },
		{ one: '{count} conversion · {rate} des visiteurs', other: '{count} conversions · {rate} des visiteurs' },
		{ one: '{count} conversión · {rate} de los visitantes', other: '{count} conversiones · {rate} de los visitantes' },
		{ one: '{count} Conversion · {rate} der Besucher', other: '{count} Conversions · {rate} der Besucher' }
	],
	'goals.typeEvent': ['event', 'événement', 'evento', 'Ereignis'],
	'goals.typePageview': ['pageview', 'page vue', 'visita', 'Seitenaufruf'],
	'goals.noConversions': [
		'No conversions in this period yet.',
		'Aucune conversion sur cette période pour le moment.',
		'Aún no hay conversiones en este periodo.',
		'In diesem Zeitraum gibt es noch keine Conversions.'
	],
	'goals.empty': [
		'No goals yet. A goal is a page or a custom event you want a conversion rate for.',
		'Aucun objectif pour le moment. Un objectif est une page ou un événement personnalisé dont vous voulez le taux de conversion.',
		'Aún no hay objetivos. Un objetivo es una página o un evento personalizado del que quieres la tasa de conversión.',
		'Noch keine Ziele. Ein Ziel ist eine Seite oder ein benutzerdefiniertes Ereignis, dessen Conversionsrate Sie wissen möchten.'
	],
	'goals.deleteAria': ['Delete {name}', 'Supprimer {name}', 'Eliminar {name}', '{name} löschen'],
	'goals.deleteTitle': ['Delete {name}?', 'Supprimer {name} ?', '¿Eliminar {name}?', '{name} löschen?'],
	'goals.thisGoal': ['this goal', 'cet objectif', 'este objetivo', 'dieses Ziel'],
	'goals.deleteBody': [
		'The goal stops being measured. Past conversions are not stored separately, so its history goes with it.',
		'L’objectif n’est plus mesuré. Les conversions passées ne sont pas stockées à part : son historique disparaît avec lui.',
		'El objetivo deja de medirse. Las conversiones pasadas no se guardan por separado, así que su historial desaparece con él.',
		'Das Ziel wird nicht mehr gemessen. Vergangene Conversions werden nicht gesondert gespeichert, daher geht sein Verlauf mit verloren.'
	],
	'goals.deleteConfirm': ['Delete goal', 'Supprimer l’objectif', 'Eliminar objetivo', 'Ziel löschen'],
	'goals.deleted': ['Goal deleted.', 'Objectif supprimé.', 'Objetivo eliminado.', 'Ziel gelöscht.'],
	'goals.added': ['Goal added.', 'Objectif ajouté.', 'Objetivo añadido.', 'Ziel hinzugefügt.'],
	'goals.saveFailed': [
		'Could not save the goal.',
		'Impossible d’enregistrer l’objectif.',
		'No se pudo guardar el objetivo.',
		'Das Ziel konnte nicht gespeichert werden.'
	],
	'goals.notSaved': ['Not saved', 'Non enregistré', 'No se guardó', 'Nicht gespeichert'],
	'goals.newTitle': ['New goal', 'Nouvel objectif', 'Nuevo objetivo', 'Neues Ziel'],
	'goals.newDescription': [
		'A goal turns a page or an event into a conversion rate.',
		'Un objectif transforme une page ou un événement en taux de conversion.',
		'Un objetivo convierte una página o un evento en una tasa de conversión.',
		'Ein Ziel macht aus einer Seite oder einem Ereignis eine Conversionsrate.'
	],
	'goals.nameHelper': [
		'How it appears in the goals list.',
		'Son nom dans la liste des objectifs.',
		'Cómo aparece en la lista de objetivos.',
		'So erscheint es in der Zielliste.'
	],
	'goals.namePlaceholder': ['Signup', 'Inscription', 'Registro', 'Registrierung'],
	'goals.type': ['Type', 'Type', 'Tipo', 'Typ'],
	'goals.typePageviewOption': [
		'Pageview — someone reaches a page',
		'Page vue : quelqu’un atteint une page',
		'Visita: alguien llega a una página',
		'Seitenaufruf – jemand erreicht eine Seite'
	],
	'goals.typeEventOption': [
		'Custom event — the script reports one',
		'Événement personnalisé : le script en signale un',
		'Evento personalizado: el script informa de uno',
		'Benutzerdefiniertes Ereignis – das Script meldet eines'
	],
	'goals.path': ['Path', 'Chemin', 'Ruta', 'Pfad'],
	'goals.match': ['Match', 'Correspondance', 'Coincidencia', 'Übereinstimmung'],
	'goals.matchExact': ['Exactly this path', 'Exactement ce chemin', 'Exactamente esta ruta', 'Genau dieser Pfad'],
	'goals.matchStarts': ['Starts with it', 'Commence par', 'Empieza por', 'Beginnt mit'],
	'goals.matchContains': ['Contains it', 'Contient', 'Contiene', 'Enthält'],
	'goals.eventName': ['Event name', 'Nom de l’événement', 'Nombre del evento', 'Ereignisname'],
	'goals.eventHelper': [
		'The string your script passes to the tracker.',
		'La chaîne que votre script transmet au traceur.',
		'La cadena que tu script pasa al rastreador.',
		'Die Zeichenfolge, die Ihr Script an den Tracker übergibt.'
	],

	'perf.title': ['Page load', 'Chargement des pages', 'Carga de página', 'Ladezeit'],
	'perf.text': [
		{ one: 'Averaged over {count} sample, in milliseconds.', other: 'Averaged over {count} samples, in milliseconds.' },
		{ one: 'Moyenne sur {count} échantillon, en millisecondes.', other: 'Moyenne sur {count} échantillons, en millisecondes.' },
		{ one: 'Promedio de {count} muestra, en milisegundos.', other: 'Promedio de {count} muestras, en milisegundos.' },
		{ one: 'Durchschnitt aus {count} Stichprobe, in Millisekunden.', other: 'Durchschnitt aus {count} Stichproben, in Millisekunden.' }
	],
	'perf.series': ['Milliseconds', 'Millisecondes', 'Milisegundos', 'Millisekunden'],
	'perf.unit': ['{value} ms', '{value} ms', '{value} ms', '{value} ms'],

	'map.title': ['Where they are', 'D’où ils viennent', 'Dónde están', 'Woher sie kommen'],
	'map.text': ['Visitors by country.', 'Visiteurs par pays.', 'Visitantes por país.', 'Besucher nach Land.'],
	'map.unknown': ['Unknown', 'Inconnu', 'Desconocido', 'Unbekannt'],
	'map.country': ['Country', 'Pays', 'País', 'Land'],
	'map.zoomIn': ['Zoom in', 'Zoom avant', 'Acercar', 'Vergrößern'],
	'map.zoomOut': ['Zoom out', 'Zoom arrière', 'Alejar', 'Verkleinern'],
	'map.reset': ['Reset view', 'Réinitialiser la vue', 'Restablecer vista', 'Ansicht zurücksetzen'],

	'content.title': ['Content', 'Contenu', 'Contenido', 'Inhalte'],
	'campaigns.title': ['Campaigns and events', 'Campagnes et événements', 'Campañas y eventos', 'Kampagnen und Ereignisse'],
	'block.pages': ['Top pages', 'Pages les plus vues', 'Páginas principales', 'Top-Seiten'],
	'block.referrers': ['Top referrers', 'Sources principales', 'Orígenes principales', 'Top-Referrer'],
	'block.browsers': ['Browsers', 'Navigateurs', 'Navegadores', 'Browser'],
	'block.os': ['Operating systems', 'Systèmes d’exploitation', 'Sistemas operativos', 'Betriebssysteme'],
	'block.entry': ['Entry pages', 'Pages d’entrée', 'Páginas de entrada', 'Einstiegsseiten'],
	'block.exit': ['Exit pages', 'Pages de sortie', 'Páginas de salida', 'Ausstiegsseiten'],
	'block.countries': ['Countries', 'Pays', 'Países', 'Länder'],
	'block.devices': ['Devices', 'Appareils', 'Dispositivos', 'Geräte'],
	'block.screens': ['Screens', 'Écrans', 'Pantallas', 'Bildschirme'],
	'block.utmSources': ['UTM sources', 'Sources UTM', 'Fuentes UTM', 'UTM-Quellen'],
	'block.utmMediums': ['UTM mediums', 'Médias UTM', 'Medios UTM', 'UTM-Medien'],
	'block.utmCampaigns': ['UTM campaigns', 'Campagnes UTM', 'Campañas UTM', 'UTM-Kampagnen'],
	'block.events': ['Custom events', 'Événements personnalisés', 'Eventos personalizados', 'Benutzerdefinierte Ereignisse'],

	'public.notFoundTitle': ['Dashboard not found', 'Tableau de bord introuvable', 'Panel no encontrado', 'Dashboard nicht gefunden'],
	'public.notFoundText': [
		'This link has been revoked, or it never existed. Ask whoever shared it for a new one.',
		'Ce lien a été révoqué ou n’a jamais existé. Demandez-en un nouveau à la personne qui l’a partagé.',
		'Este enlace fue revocado o nunca existió. Pide uno nuevo a quien lo compartió.',
		'Dieser Link wurde widerrufen oder hat nie existiert. Bitten Sie die Person, die ihn geteilt hat, um einen neuen.'
	],
	'public.powered': ['Powered by Vision', 'Propulsé par Vision', 'Con tecnología de Vision', 'Bereitgestellt von Vision']
} satisfies Record<string, Row>;
