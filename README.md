# Marea

Marea es una base autohospedada para gestionar redes sociales desde un solo lugar: calendario, editor de publicaciones, biblioteca de medios, bandeja de mensajes y herramientas de IA. La IA prepara borradores; la persona revisa y confirma antes de publicar o responder.

Marea parte de [OpenPost](https://github.com/getopenpost/openpost), que se distribuye bajo AGPL-3.0. Consulta [LICENSE](LICENSE) y conserva los avisos de licencia al modificar o redistribuir el proyecto.

## Qué incluye esta base

- Aplicación web y API de OpenPost con la marca Marea.
- Opción de conectar un servidor local compatible con OpenAI, como `llama.cpp`, para generar texto sin Ollama ni claves de pago.
- Adaptador de imágenes a ComfyUI para generar y editar imágenes en infraestructura propia.
- Despliegue Docker Compose con Caddy para HTTPS.
- Servicios de IA aislados en la red interna de Docker.

## Requisitos

- VPS con Docker Compose v2.
- Dominio apuntado al VPS y puertos 80/443 disponibles para HTTPS.
- Para texto: RAM y CPU suficientes para el modelo GGUF escogido. El ejemplo usa Qwen2.5 3B cuantizado.
- Para imágenes: ComfyUI, un checkpoint compatible y, de preferencia, una GPU. La generación en CPU puede ser muy lenta.
- Espacio de disco para modelos y medios. Los pesos de IA no se incluyen en el repositorio.

## Preparar el servidor

1. Copia `.env.example` como `.env`; crea secretos únicos y completa `MAREA_DOMAIN`.
2. Descarga un modelo GGUF compatible en `models/llm/` (por ejemplo, Qwen2.5 3B Instruct Q4_K_M desde Hugging Face).
3. Descarga e instala un checkpoint para ComfyUI en `models/comfyui/checkpoints/`. El nombre debe coincidir con `MAREA_COMFYUI_CHECKPOINT`.
4. Desde la raíz del repositorio, ejecuta `docker compose -f deploy/marea/compose.yml up -d --build`.
5. Abre el dominio, crea la cuenta administradora y conecta las redes desde Ajustes.
6. Cuando la cuenta esté creada, desactiva el registro público cambiando `OPENPOST_DISABLE_REGISTRATIONS=true` y reinicia el servicio web.

El compose publica solo la web y Caddy. `llama.cpp` y ComfyUI quedan dentro de la red privada de Docker. Nunca publiques sus puertos directamente en Internet.

## IA local

El archivo `.env.example` muestra las variables del servicio de texto. El endpoint compatible con OpenAI se mantiene privado y recibe una clave local ficticia, no una cuenta de proveedor. Para cambiar el modelo, ajusta la ruta del archivo y el alias en `deploy/marea/compose.yml`.

La generación de imágenes es opcional: instala los pesos de ComfyUI y configura su nombre de archivo. Los pesos tienen licencias propias; revísalas antes de uso comercial. Sin GPU, esta función puede ser impráctica en un VPS pequeño.

## Redes sociales y límites de acceso

Marea utiliza las integraciones disponibles en OpenPost y las APIs oficiales de las plataformas. Instagram requiere una cuenta profesional y una aplicación de Meta con permisos aprobados para funciones como comentarios y mensajes. TikTok limita la publicación directa de aplicaciones no auditadas y no ofrece acceso general a funciones como mensajes privados o cambio de perfil. Por eso, la disponibilidad exacta depende de los permisos que apruebe cada plataforma; no se puede prometer control total ni saltarse sus APIs.

La programación de contenido solo debe activarse después de revisar el borrador y confirmar la acción en Marea. Mantén desactivados los flujos automáticos hasta que la persona administradora los haya revisado.

## Desarrollo

Consulta la documentación de OpenPost y sus avisos AGPL antes de desarrollar o redistribuir cambios. El proyecto upstream requiere Bun y Go para construir la aplicación. El VPS requiere Docker Compose; no hace falta instalar Ollama.
