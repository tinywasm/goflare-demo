# Setup: D1 Binding & Custom Domain

Este documento describe los pasos manuales necesarios en el dashboard de Cloudflare para que el pipeline de CI/CD y el demo funcionen correctamente.

## 1. D1 Database Binding

Para que las Pages Functions puedan acceder a la base de datos D1:

1. Ve a **Cloudflare Dashboard** → **Workers & Pages**.
2. Selecciona el proyecto `goflare-demo`.
3. Ve a **Settings** → **Functions**.
4. Baja hasta **D1 database bindings**.
5. Haz clic en **Add binding**.
6. Configura:
   - **Variable name**: `DB`
   - **D1 database**: Selecciona tu base de datos (ej. `goflare-db`).
7. Haz clic en **Save**.

*Nota: Repite esto tanto en "Production" como en "Preview" si deseas que ambos entornos funcionen.*

## 2. Dominio Personalizado

Para configurar `goflare-demo.tinywasm.app`:

1. En el proyecto de Pages, ve a **Custom domains**.
2. Haz clic en **Add custom domain**.
3. Introduce `goflare-demo.tinywasm.app`.
4. Cloudflare detectará que el dominio está en la misma cuenta y ofrecerá configurar el DNS automáticamente.
5. Asegúrate de que el registro sea un **CNAME** apuntando a `goflare-demo.pages.dev` con el **Proxy activado** (nube naranja).

## 3. Variables de Entorno para CI (GitHub Actions)

Asegúrate de tener estos secrets y variables en GitHub:

- **Secrets**:
  - `CLOUDFLARE_API_TOKEN`: Token con permisos para editar Pages y consultar D1.
  - `CLOUDFLARE_ACCOUNT_ID`: Tu ID de cuenta de Cloudflare.
- **Variables**:
  - `D1_DATABASE_ID`: El UUID de la base de datos D1.
