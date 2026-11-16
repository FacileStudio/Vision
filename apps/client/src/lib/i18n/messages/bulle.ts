import type { Row } from '../row';

/** The Bulle panel on the site page: anomalies, their explanation, and ask-your-data. */
export const bulle = {
	'bulle.label': ['Insights', 'Analyses', 'Análisis', 'Einblicke'],
	'bulle.anomalies.heading': ['Anomalies', 'Anomalies', 'Anomalías', 'Auffälligkeiten'],
	'bulle.anomalies.intro': [
		'Days when traffic left its usual range, compared with the same weekday of the previous weeks. No AI finds them.',
		'Jours où le trafic est sorti de sa plage habituelle, comparés au même jour de la semaine des semaines précédentes. Aucune IA ne les trouve.',
		'Días en que el tráfico salió de su rango habitual, comparados con el mismo día de la semana en las semanas anteriores. Ninguna IA los encuentra.',
		'Tage, an denen der Traffic seinen üblichen Bereich verlassen hat, verglichen mit demselben Wochentag der Vorwochen. Keine KI findet sie.'
	],
	'bulle.anomalies.none': [
		'No anomalies in the last 30 days.',
		'Aucune anomalie ces 30 derniers jours.',
		'Ninguna anomalía en los últimos 30 días.',
		'Keine Auffälligkeiten in den letzten 30 Tagen.'
	],
	'bulle.anomalies.loading': ['Loading anomalies', 'Chargement des anomalies', 'Cargando anomalías', 'Auffälligkeiten werden geladen'],
	'bulle.metric.visitors': ['Visitors', 'Visiteurs', 'Visitantes', 'Besucher'],
	'bulle.metric.pageviews': ['Pageviews', 'Pages vues', 'Páginas vistas', 'Seitenaufrufe'],
	'bulle.metric.goal_conversions': ['Goal: {goal}', 'Objectif : {goal}', 'Objetivo: {goal}', 'Ziel: {goal}'],
	'bulle.metric.top_referrer_share': [
		'Top referrer share',
		'Part du premier référent',
		'Cuota del primer referente',
		'Anteil der Top-Quelle'
	],
	'bulle.direction.up': ['Above usual', 'Au-dessus de l’habituel', 'Por encima de lo habitual', 'Über dem Üblichen'],
	'bulle.direction.down': ['Below usual', 'En dessous de l’habituel', 'Por debajo de lo habitual', 'Unter dem Üblichen'],
	'bulle.compare': [
		'{value} against {baseline} usual ({change})',
		'{value} contre {baseline} d’habitude ({change})',
		'{value} frente a {baseline} habitual ({change})',
		'{value} gegenüber üblichen {baseline} ({change})'
	],
	'bulle.compareNew': [
		'{value} against {baseline} usual',
		'{value} contre {baseline} d’habitude',
		'{value} frente a {baseline} habitual',
		'{value} gegenüber üblichen {baseline}'
	],
	'bulle.movers': ['What moved', 'Ce qui a bougé', 'Qué se movió', 'Was sich bewegt hat'],
	'bulle.dim.referrers': ['Referrers', 'Référents', 'Referentes', 'Quellen'],
	'bulle.dim.paths': ['Pages', 'Pages', 'Páginas', 'Seiten'],
	'bulle.dim.countries': ['Countries', 'Pays', 'Países', 'Länder'],
	'bulle.dim.devices': ['Devices', 'Appareils', 'Dispositivos', 'Geräte'],
	'bulle.direct': ['(direct)', '(direct)', '(directo)', '(direkt)'],
	'bulle.unknown': ['(not recorded)', '(non renseigné)', '(sin registrar)', '(nicht erfasst)'],
	'bulle.mover': ['{count} views, {delta} against usual', '{count} vues, {delta} par rapport à l’habituel', '{count} vistas, {delta} frente a lo habitual', '{count} Aufrufe, {delta} gegenüber dem Üblichen'],
	'bulle.explain': ['Explain', 'Expliquer', 'Explicar', 'Erklären'],
	'bulle.explain.aria': [
		'Explain the {metric} anomaly of {day}',
		'Expliquer l’anomalie {metric} du {day}',
		'Explicar la anomalía de {metric} del {day}',
		'Auffälligkeit {metric} vom {day} erklären'
	],
	'bulle.explain.causes': ['Likely causes', 'Causes probables', 'Causas probables', 'Wahrscheinliche Ursachen'],
	'bulle.explain.check': ['What to check', 'À vérifier', 'Qué comprobar', 'Was zu prüfen ist'],
	'bulle.confidence.low': ['low confidence', 'confiance faible', 'confianza baja', 'geringe Sicherheit'],
	'bulle.confidence.medium': ['medium confidence', 'confiance moyenne', 'confianza media', 'mittlere Sicherheit'],
	'bulle.confidence.high': ['high confidence', 'confiance élevée', 'confianza alta', 'hohe Sicherheit'],
	'bulle.ask.heading': ['Ask your data', 'Interroger vos données', 'Pregunta a tus datos', 'Fragen Sie Ihre Daten'],
	'bulle.ask.label': ['Your question', 'Votre question', 'Tu pregunta', 'Ihre Frage'],
	'bulle.ask.placeholder': [
		'Why did visitors drop last week?',
		'Pourquoi les visiteurs ont-ils baissé la semaine dernière ?',
		'¿Por qué bajaron los visitantes la semana pasada?',
		'Warum sind die Besucher letzte Woche zurückgegangen?'
	],
	'bulle.ask.hint': [
		'Answers come from this site’s numbers only. Nothing is changed.',
		'Les réponses viennent uniquement des chiffres de ce site. Rien n’est modifié.',
		'Las respuestas salen solo de las cifras de este sitio. No se cambia nada.',
		'Die Antworten stammen nur aus den Zahlen dieser Website. Es wird nichts geändert.'
	],
	'bulle.ask.submit': ['Ask', 'Demander', 'Preguntar', 'Fragen'],
	'bulle.ask.answer': ['Answer', 'Réponse', 'Respuesta', 'Antwort'],
	'bulle.ask.needsApproval': [
		'Bulle is waiting for your approval. Continue in Bulle to answer it.',
		'Bulle attend votre accord. Continuez dans Bulle pour répondre.',
		'Bulle espera tu aprobación. Continúa en Bulle para responder.',
		'Bulle wartet auf Ihre Freigabe. Machen Sie in Bulle weiter, um zu antworten.'
	],
	'bulle.ask.failed': [
		'The run did not finish. Try again, or continue in Bulle.',
		'L’exécution ne s’est pas terminée. Réessayez, ou continuez dans Bulle.',
		'La ejecución no terminó. Inténtalo de nuevo o continúa en Bulle.',
		'Die Ausführung wurde nicht beendet. Versuchen Sie es erneut oder machen Sie in Bulle weiter.'
	],
	'bulle.ask.continue': ['Continue in Bulle', 'Continuer dans Bulle', 'Continuar en Bulle', 'In Bulle weitermachen'],
	'bulle.error.load': [
		'Could not load the anomalies.',
		'Impossible de charger les anomalies.',
		'No se pudieron cargar las anomalías.',
		'Die Auffälligkeiten konnten nicht geladen werden.'
	],
	'bulle.error.unreachable': [
		'Bulle is unreachable for now. Try again in a moment.',
		'Bulle est injoignable pour le moment. Réessayez dans un instant.',
		'Bulle no responde por ahora. Inténtalo en un momento.',
		'Bulle ist gerade nicht erreichbar. Versuchen Sie es gleich noch einmal.'
	],
	'bulle.error.capped': [
		'You reached Bulle’s limit for now. Try again later.',
		'Vous avez atteint la limite de Bulle pour le moment. Réessayez plus tard.',
		'Has alcanzado el límite de Bulle por ahora. Inténtalo más tarde.',
		'Sie haben Bulles Limit vorerst erreicht. Versuchen Sie es später erneut.'
	],
	'bulle.error.notLinked': [
		'Reconnect your Facile account to use Bulle, or pick a space that Registre manages.',
		'Reconnectez votre compte Facile pour utiliser Bulle, ou choisissez un espace géré par Registre.',
		'Vuelve a conectar tu cuenta Facile para usar Bulle, o elige un espacio gestionado por Registre.',
		'Verbinden Sie Ihr Facile-Konto erneut oder wählen Sie einen von Registre verwalteten Bereich.'
	],
	'bulle.upsell': [
		'Explaining anomalies and answering questions need a plan with AI access. Finding anomalies stays free.',
		'Expliquer les anomalies et répondre aux questions demande un forfait avec accès à l’IA. Repérer les anomalies reste gratuit.',
		'Explicar anomalías y responder preguntas requiere un plan con acceso a IA. Detectar anomalías sigue siendo gratis.',
		'Auffälligkeiten erklären und Fragen beantworten erfordert einen Tarif mit KI-Zugang. Das Finden bleibt kostenlos.'
	]
} satisfies Record<string, Row>;
