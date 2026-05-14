import type { RequestHandler } from './$types';

const API_URL = process.env.API_URL ?? 'http://localhost:4000';

export const GET: RequestHandler = async ({ params, url }) => {
	const token = url.searchParams.get('token') ?? '';
	const upstream = `${API_URL}/events/${params.siteId}/live?token=${encodeURIComponent(token)}`;

	const res = await fetch(upstream, {
		headers: { 'Accept': 'text/event-stream' }
	});

	if (!res.ok || !res.body) {
		const body = await res.text();
		return new Response(body, {
			status: res.status,
			headers: { 'Content-Type': 'application/json' }
		});
	}

	const stream = new ReadableStream({
		async start(controller) {
			const reader = res.body!.getReader();
			try {
				while (true) {
					const { done, value } = await reader.read();
					if (done) break;
					controller.enqueue(value);
				}
			} catch {
			} finally {
				controller.close();
			}
		},
		cancel() {
			res.body!.cancel();
		}
	});

	return new Response(stream, {
		headers: {
			'Content-Type': 'text/event-stream',
			'Cache-Control': 'no-cache',
			'Connection': 'keep-alive',
			'X-Accel-Buffering': 'no'
		}
	});
};
