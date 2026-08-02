# Kaizen Backend

## Objetivo

Backend de una aplicación de finanzas personales.

Tecnologías

- Go
- Firestore
- Cloud Run
- REST API

## Arquitectura

Seguimos Clean Architecture.

```
handler
↓

usecase
↓

repository
↓

firestore
```

No acceder a Firestore directamente desde handlers.

Toda la lógica de negocio debe vivir en UseCases.

## Estilo

- Código simple (KISS)
- Sin reflexión
- Sin frameworks pesados
- Interfaces pequeñas
- Métodos cortos
- Errores explícitos

## API

Toda funcionalidad nueva debe exponerse mediante REST.

Si hace falta modificar el contrato REST, actualizar Swagger.

## Firestore

Cada colección tiene su propio Repository.

Nunca acceder a Firestore fuera del Repository.

## Frontend

El frontend consume únicamente la API REST.

Nunca generar HTML desde Go.

## Objetivo del proyecto

Responder tres preguntas:

- ¿Cuánto dinero puedo gastar hoy?
- ¿Cuántos días faltan para mi próximo ingreso?
- ¿Voy bien o voy mal?

Toda nueva funcionalidad debe aportar valor a alguna de estas preguntas.