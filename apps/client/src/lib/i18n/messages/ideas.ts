import type { Row } from '../row';

/** The Ideas page: suggestions Bulle makes about a space's sites, run in place. */
export const ideas = {
	'ideas.title': ['Ideas · Vision', 'Idées · Vision', 'Ideas · Vision', 'Ideen · Vision'],
	'ideas.description': [
		'Bulle looks at your sites and suggests what to watch, check or set up, such as an anomaly to look into or a weekly digest. Run an idea and it is done here.',
		'Bulle regarde vos sites et propose quoi surveiller, vérifier ou mettre en place, comme une anomalie à examiner ou un résumé hebdomadaire. Lancez une idée, elle se fait ici.',
		'Bulle revisa tus sitios y propone qué vigilar, comprobar o configurar, como una anomalía que revisar o un resumen semanal. Lanza una idea y se hace aquí.',
		'Bulle sieht sich Ihre Websites an und schlägt vor, was zu beobachten, zu prüfen oder einzurichten ist, etwa eine Auffälligkeit oder eine Wochenübersicht. Starten Sie eine Idee, sie wird hier erledigt.'
	],
	'ideas.space': ['Space', 'Espace', 'Espacio', 'Bereich'],
	'ideas.dismissed': ['Idea dismissed', 'Idée ignorée', 'Idea descartada', 'Idee verworfen'],
	'ideas.approval.summary': [
		'Bulle wants to run "{tool}". Check it before it goes ahead.',
		'Bulle veut lancer « {tool} ». Vérifiez avant de continuer.',
		'Bulle quiere ejecutar «{tool}». Revísalo antes de seguir.',
		'Bulle möchte „{tool}“ ausführen. Prüfen Sie es, bevor es weitergeht.'
	],
	'ideas.error.load': [
		'Could not load the ideas.',
		'Impossible de charger les idées.',
		'No se pudieron cargar las ideas.',
		'Die Ideen konnten nicht geladen werden.'
	],
	'ideas.error.run': [
		'The run did not finish.',
		'L’exécution ne s’est pas terminée.',
		'La ejecución no terminó.',
		'Die Ausführung wurde nicht beendet.'
	],
	'ideas.error.busy': [
		'Bulle is already working on a batch. Try again in a moment.',
		'Bulle prépare déjà un lot. Réessayez dans un instant.',
		'Bulle ya está preparando un lote. Inténtalo en un momento.',
		'Bulle bereitet bereits einen Stapel vor. Versuchen Sie es gleich noch einmal.'
	],
	'ideas.error.gone': [
		'That idea is gone.',
		'Cette idée n’existe plus.',
		'Esa idea ya no existe.',
		'Diese Idee gibt es nicht mehr.'
	],
	'ideas.example.drop.title': [
		'Visitors dropped on Tuesday',
		'Les visiteurs ont baissé mardi',
		'Los visitantes bajaron el martes',
		'Die Besucher gingen am Dienstag zurück'
	],
	'ideas.example.drop.why': [
		'Traffic was 40% under the usual Tuesday. Bulle looks at what moved.',
		'Le trafic était 40 % sous le mardi habituel. Bulle regarde ce qui a bougé.',
		'El tráfico estuvo un 40 % por debajo del martes habitual. Bulle mira qué se movió.',
		'Der Traffic lag 40 % unter einem üblichen Dienstag. Bulle prüft, was sich bewegt hat.'
	],
	'ideas.example.digest.title': [
		'Your weekly digest',
		'Votre résumé hebdomadaire',
		'Tu resumen semanal',
		'Ihre Wochenübersicht'
	],
	'ideas.example.digest.why': [
		'The week against the one before, in a few lines.',
		'La semaine face à la précédente, en quelques lignes.',
		'La semana frente a la anterior, en pocas líneas.',
		'Die Woche im Vergleich zur Vorwoche, in wenigen Zeilen.'
	],
	'ideas.example.goal.title': [
		'A goal worth adding',
		'Un objectif à ajouter',
		'Un objetivo que añadir',
		'Ein Ziel, das sich lohnt'
	],
	'ideas.example.goal.why': [
		'Visitors keep reaching /thanks and no goal counts them.',
		'Des visiteurs atteignent /merci et aucun objectif ne les compte.',
		'Los visitantes llegan a /gracias y ningún objetivo los cuenta.',
		'Besucher erreichen /danke und kein Ziel zählt sie.'
	]
} satisfies Record<string, Row>;
