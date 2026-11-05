<script lang="ts">
	import { CalendarDays, Check, Image, MessageCircle, Sparkles } from '@lucide/svelte';
	import LanguageSwitcher from '$lib/components/language-switcher.svelte';
	import Logo from '$lib/components/Logo.svelte';
	import { Button } from '$lib/components/ui/button';

	const features = [
		{ icon: Sparkles, title: 'Ideas y publicaciones', detail: 'Planificá, redactá y adaptá contenido para cada red.' },
		{ icon: Image, title: 'Tu galería y editor', detail: 'Organizá fotos y videos, prepará piezas y revisalas antes de usarlas.' },
		{ icon: MessageCircle, title: 'Bandeja unificada', detail: 'Revisá y respondé comentarios y mensajes cuando la plataforma lo permita.' },
		{ icon: CalendarDays, title: 'Calendario de contenido', detail: 'Mantené tus ideas, borradores y publicaciones en un solo lugar.' }
	];
</script>

<svelte:head>
	<title>Marea — tu CM con IA</title>
	<meta name="description" content="Planificá, prepará y gestioná el contenido de tus redes con Marea. Vos aprobás antes de cada envío." />
</svelte:head>

<div class="min-h-screen bg-background text-foreground selection:bg-primary/20 selection:text-foreground">
	<header class="border-b">
		<div class="mx-auto flex min-h-16 w-full max-w-6xl items-center justify-between gap-4 px-4 sm:px-6">
			<a href="/" class="inline-flex min-h-11 items-center gap-2 rounded-md" aria-label="Marea">
				<Logo width={32} height={32} decorative />
				<span class="font-brand text-base font-semibold tracking-[-0.02em]">Marea</span>
			</a>
			<div class="flex items-center gap-2">
				<Button href="/login" variant="ghost" size="sm">Ingresar</Button>
				<LanguageSwitcher compact />
			</div>
		</div>
	</header>

	<main class="mx-auto w-full max-w-6xl px-4 sm:px-6">
		<section class="py-16 sm:py-24" aria-labelledby="marea-title">
			<div class="max-w-3xl">
				<div class="mb-5 inline-flex items-center gap-2 rounded-full border bg-muted/40 px-3 py-1.5 text-sm text-muted-foreground">
					<Sparkles class="size-4 text-primary" aria-hidden="true" />
					<span>Tu community manager, con vos al mando</span>
				</div>
				<h1 id="marea-title" data-theme-type="display" class="text-balance">Tus redes en movimiento. Cada envío, con tu aprobación.</h1>
				<p class="mt-6 max-w-2xl text-base leading-7 text-muted-foreground sm:text-lg">
					Marea te ayuda a planificar, crear y organizar contenido, revisar la actividad de tus cuentas y preparar respuestas. Nada se publica ni se envía hasta que vos lo confirmás.
				</p>
				<div class="mt-8 flex flex-wrap gap-3">
					<Button href="/register" size="lg">Empezar</Button>
					<Button href="/login" variant="outline" size="lg">Ya tengo una cuenta</Button>
				</div>
				<p class="mt-4 max-w-xl text-sm leading-6 text-muted-foreground">
					Marea se instala en tu servidor. Las conexiones disponibles dependen de los permisos que cada red conceda a la aplicación.
				</p>
			</div>
		</section>

		<section class="border-t py-12 sm:py-16" aria-labelledby="marea-features">
			<h2 id="marea-features" class="text-2xl font-semibold tracking-tight">Un lugar para llevar el día a día de tus redes</h2>
			<ul class="mt-8 grid gap-7 sm:grid-cols-2 lg:grid-cols-4">
				{#each features as feature (feature.title)}
					{@const Icon = feature.icon}
					<li>
						<Icon class="size-5 text-primary" strokeWidth={1.8} aria-hidden="true" />
						<h3 class="mt-3 text-base font-semibold">{feature.title}</h3>
						<p class="mt-2 text-sm leading-6 text-muted-foreground">{feature.detail}</p>
					</li>
				{/each}
			</ul>
		</section>

		<section class="border-t py-10" aria-label="Aprobación manual">
			<div class="flex max-w-2xl items-start gap-3">
				<Check class="mt-0.5 size-5 shrink-0 text-primary" aria-hidden="true" />
				<p class="text-sm leading-6 text-muted-foreground">
					La IA prepara borradores y sugerencias. Vos decidís qué publicar, qué responder y qué cambiar.
				</p>
			</div>
		</section>
	</main>

	<footer class="border-t">
		<div class="mx-auto flex w-full max-w-6xl flex-col gap-2 px-4 py-6 text-sm text-muted-foreground sm:flex-row sm:items-center sm:justify-between sm:px-6">
			<span>Marea — software libre para gestionar redes.</span>
			<a class="min-h-11 self-start underline-offset-4 hover:text-foreground hover:underline sm:self-center" href="https://github.com/dcorp-sas/MAREA">Código fuente en GitHub</a>
		</div>
	</footer>
</div>
