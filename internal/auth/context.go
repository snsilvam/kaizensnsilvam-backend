// Package auth expone la identidad del usuario autenticado a través del
// context.Context.
//
// El middleware HTTP escribe acá el UID ya verificado, y los handlers y casos
// de uso lo leen desde el context.Context que de todas formas reciben. Así el
// dominio no depende ni de Gin ni del SDK de Firebase: sólo de este paquete y
// de la stdlib.
package auth

import "context"

// uidKey es la clave del valor en el contexto. Es un tipo privado, no un
// string, por dos razones: nadie fuera de este paquete puede escribir ni pisar
// el UID, y no puede colisionar con las claves de otros paquetes.
type uidKey struct{}

// WithUID devuelve una copia de ctx que lleva el UID del usuario autenticado.
// Sólo lo usa el middleware de autenticación.
func WithUID(ctx context.Context, uid string) context.Context {
	return context.WithValue(ctx, uidKey{}, uid)
}

// UID devuelve el UID de Firebase del usuario autenticado.
//
// El segundo valor distingue los dos casos: true es una petición autenticada
// con UID; false es una petición que no pasó por el middleware de
// autenticación, es decir, sin usuario. Un UID vacío cuenta como no
// autenticado porque no identifica a nadie.
func UID(ctx context.Context) (string, bool) {
	uid, ok := ctx.Value(uidKey{}).(string)
	if !ok || uid == "" {
		return "", false
	}
	return uid, true
}
