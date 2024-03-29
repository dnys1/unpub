import 'package:ngdart/angular.dart';
import 'package:ngrouter/ngrouter.dart';
import 'package:unpub_web/api/models.dart';
import 'package:unpub_web/app_service.dart';
import 'routes.dart';

@Component(
  selector: 'home',
  templateUrl: 'home_component.html',
  directives: [routerDirectives, coreDirectives],
  exports: [RoutePaths],
)
class HomeComponent implements OnActivate {
  final AppService appService;

  ListApi? data;

  HomeComponent(this.appService);

  @override
  void onActivate(RouterState? previous, RouterState current) async {
    appService.setLoading(true);
    data = await appService.fetchPackages(size: 15);
    appService.setLoading(false);
  }

  getDetailUrl(ListApiPackage package) {
    return RoutePaths.detail.toUrl(parameters: {'name': package.name});
  }
}
