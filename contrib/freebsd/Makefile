PORTVERSION=	0.4.3
CATEGORIES=	sysutils

MAINTAINER=	Hermes J. Rodríguez Azuaje <hejeroaz@gmail.com>
COMMENT=	Privacy-first web analytics collector

NO_BUILD=	yes
NO_ARCH=	yes

DISTNAME=	kiko-${PORTVERSION}

MASTER_SITES=	https://github.com/hrodrig/kiko/releases/download/v${PORTVERSION}/
DISTFILES=	kiko_${PORTVERSION}_freebsd_amd64.tar.gz

PLIST=	${.CURDIR}/pkg-plist

do-install:
	${INSTALL_PROGRAM} ${WRKSRC}/kiko ${STAGEDIR}${PREFIX}/bin/
	${INSTALL_MAN} ${WRKSRC}/kiko.1 ${STAGEDIR}${PREFIX}/man/man1/
	${INSTALL_DATA} ${WRKSRC}/LICENSE ${STAGEDIR}${PREFIX}/share/doc/kiko/

.include <bsd.port.mk>
