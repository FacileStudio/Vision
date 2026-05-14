<script lang="ts">
	import { goto } from '$app/navigation';
	import { api, setToken } from '$lib';

	let email = $state('');
	let password = $state('');
	let error = $state('');
	let isRegister = $state(false);

	async function submit() {
		error = '';
		try {
			const fn = isRegister ? api.auth.register : api.auth.login;
			const res = await fn(email, password);
			setToken(res.token);
			goto('/dashboard');
		} catch (e: any) {
			error = e.message;
		}
	}
</script>

<div class="flex min-h-screen items-center justify-center">
	<div class="w-full max-w-sm space-y-6 p-8">
		<h1 class="text-2xl font-bold text-center">Vision</h1>
		<p class="text-muted-foreground text-center text-sm">
			{isRegister ? 'Create your account' : 'Sign in to your account'}
		</p>

		{#if error}
			<p class="text-destructive text-sm text-center">{error}</p>
		{/if}

		<form onsubmit={submit} class="space-y-4">
			<input
				type="email"
				placeholder="Email"
				bind:value={email}
				class="w-full rounded-md border bg-background px-3 py-2 text-sm"
				required
			/>
			<input
				type="password"
				placeholder="Password"
				bind:value={password}
				class="w-full rounded-md border bg-background px-3 py-2 text-sm"
				required
				minlength={8}
			/>
			<button
				type="submit"
				class="w-full rounded-md bg-primary px-3 py-2 text-sm text-primary-foreground"
			>
				{isRegister ? 'Register' : 'Login'}
			</button>
		</form>

		<button
			onclick={() => (isRegister = !isRegister)}
			class="w-full text-center text-sm text-muted-foreground hover:underline"
		>
			{isRegister ? 'Already have an account? Login' : "Don't have an account? Register"}
		</button>
	</div>
</div>
