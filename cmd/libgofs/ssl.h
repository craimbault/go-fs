/*
** ssl.h
** Simple String List
** array of characters or zero terminated strings
*/

#ifndef SSL_H  /* Pour éviter les déclarations multiples */
#define SSL_H

typedef struct _SSL *SSL;

SSL SSL_new(unsigned int size);
SSL SSL_del(SSL ssl);
unsigned int SSL_count(SSL ssl);
unsigned int SSL_add(SSL ssl, char *sz);
unsigned int SSL_add2(SSL ssl, char *str, unsigned int len);
char *SSL_get(SSL ssl, int index);
unsigned int SSL_len(SSL ssl, int index);

#endif  /* SSL_H */

/* eof */